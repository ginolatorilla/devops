use clap::ValueEnum;
use k8s_openapi::api::core::v1::{ConfigMap, Pod};
use k8s_openapi::{Metadata, NamespaceResourceScope, Resource};
use kube::api::ObjectMeta;
use kube::{
    api::{Api, ListParams, ResourceExt},
    Client,
};
use log::debug;

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
        .flat_map(|pod| extract_config_maps_from_pod_volumes(&pod))
        .collect();

    let _ = config_maps
        .iter()
        .filter(|cm_name| !used_config_maps.contains(&cm_name))
        .for_each(|cm_name| println!("Found unused configmap: {}.", cm_name));
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
        .filter(|p| -> bool {
            let is_orphan = p.metadata().owner_references.is_none();
            if is_orphan {
                debug!(
                    "Ownerless pod {} found in namespace {}.",
                    p.name_any(),
                    p.namespace().unwrap_or_default(),
                );
            }
            is_orphan
        })
        .collect();
    Ok(static_pods)
}

fn extract_config_maps_from_pod_volumes(pod: &Pod) -> Vec<String> {
    pod.spec.as_ref().map_or(Vec::new(), |pod_spec| {
        pod_spec.volumes.as_ref().map_or(Vec::new(), |volumes| {
            volumes
                .iter()
                .filter_map(|volume| {
                    volume
                        .config_map
                        .as_ref()
                        .and_then(|volume_source| {
                            let config_map_name = volume_source.name.as_ref();
                            debug!(
                                "Reference to config map {:?} found in volume {} of pod {} in namespace {}.",
                                config_map_name,
                                volume.name,
                                pod.name_any(),
                                pod.namespace().unwrap_or_default()
                            );
                            config_map_name
                        })
                })
                .fold(Vec::new(), |mut config_maps, config_map_name| {
                    config_maps.push(config_map_name.clone());
                    config_maps
                })
        })
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
