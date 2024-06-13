use clap::ValueEnum;
use k8s_openapi::api::core::v1::{ConfigMap, Pod};
use k8s_openapi::{Metadata, NamespaceResourceScope, Resource};
use kube::api::ObjectMeta;
use kube::{
    api::{Api, ListParams, ResourceExt},
    Client,
};
use log::{debug, info};

#[derive(ValueEnum, Debug, Clone)]
pub enum Resources {
    ConfigMap,
}

#[tokio::main()]
pub async fn kubeclean(
    resource: Resources,
    namespace: Option<String>,
) -> Result<(), Box<dyn std::error::Error>> {
    if namespace.is_none() {
        debug!("No namespace specified, will use what's in the current context.");
    }
    let client = Client::try_default().await?;
    match resource {
        Resources::ConfigMap => clean_config_maps(client, namespace.as_ref()).await,
    }
}

async fn clean_config_maps(
    client: Client,
    namespace: Option<&String>,
) -> Result<(), Box<dyn std::error::Error>> {
    let config_maps: Vec<String> = get_api::<ConfigMap>(client.clone(), namespace)
        .list(&ListParams::default())
        .await?
        .into_iter()
        .map(|cm| cm.name_any())
        .collect();

    let used_config_maps: Vec<String> = free_pods(client, namespace)
        .await?
        .into_iter()
        .flat_map(|pod| extract_config_maps_from_pod(&pod))
        .collect();

    config_maps
        .iter()
        .filter(|cm_name| !used_config_maps.contains(&cm_name))
        .for_each(|cm_name| info!("Found unused configmap: {}.", cm_name));
    Ok(())
}

async fn free_pods(
    client: Client,
    namespace: Option<&String>,
) -> Result<Vec<Pod>, Box<dyn std::error::Error>> {
    let pods: Api<Pod> = get_api(client, namespace);
    let static_pods = pods
        .list(&ListParams::default())
        .await?
        .into_iter()
        .filter(|pod| !pod.owner_references().is_empty())
        .inspect(|pod| {
            debug!(
                "Ownerless pod {} found in namespace {}.",
                pod.name_any(),
                pod.namespace().unwrap_or_default(),
            );
        })
        .collect();
    Ok(static_pods)
}

fn extract_config_maps_from_pod(pod: &Pod) -> Vec<String> {
    let mut a = extract_config_maps_from_pod_volumes(pod);
    let mut b = extract_config_maps_from_projected_volumes(pod);
    a.append(&mut b);
    a
}

fn extract_config_maps_from_pod_volumes(pod: &Pod) -> Vec<String> {
    pod.spec
        .as_ref()
        .and_then(|spec| spec.volumes.as_ref())
        .map_or(Vec::new(), |volumes| {
            volumes
                .iter()
                .filter_map(|volume| Some((&volume.name, volume.config_map.as_ref()?)))
                .filter_map(|(v, cvs)| Some((v, cvs.name.as_ref()?)))
                .inspect(|(v, cm)| {
                    debug!(
                        "Reference to config map {} found in volume {} in pod {} in namespace {}.",
                        cm,
                        v,
                        pod.name_any(),
                        pod.namespace().unwrap_or_default()
                    )
                })
                .map(|(_, cm)| cm.clone())
                .collect()
        })
}

fn extract_config_maps_from_projected_volumes(pod: &Pod) -> Vec<String> {
    pod.spec
        .as_ref()
        .and_then(|spec| spec.volumes.as_ref())
        .map_or(Vec::new(), |volumes| {
            volumes
                .iter()
                .filter_map(|volume| Some((&volume.name, volume.projected.as_ref()?)))
                .filter_map(|(v, pvs)| Some((v, pvs.sources.as_ref()?)))
                .flat_map(|(v, vps)| {
                    vps.iter()
                        .filter_map(move |vp| Some((v, vp.config_map.as_ref()?)))
                })
                .filter_map(|(v, cmp)| Some((v, cmp.name.as_ref()?)))
                .inspect(|(v, cm)| {
                    debug!(
                        "Reference to config map {} found in volume {} in pod {} in namespace {}.",
                        cm,
                        v,
                        pod.name_any(),
                        pod.namespace().unwrap_or_default()
                    )
                })
                .map(|(_, cm)| cm.clone())
                .collect()
        })
}

fn get_api<R: Resource<Scope = NamespaceResourceScope> + Metadata<Ty = ObjectMeta>>(
    client: Client,
    namespace: Option<&String>,
) -> Api<R> {
    match namespace {
        Some(n) => Api::namespaced(client, n),
        None => Api::default_namespaced(client),
    }
}
