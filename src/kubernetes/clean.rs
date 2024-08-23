use std::borrow::Cow;
use std::collections::HashSet;
use std::fmt::Debug;

use clap::{ArgAction, Args, ValueEnum};
use k8s_openapi::api::apps::v1::{DaemonSet, Deployment, ReplicaSet, StatefulSet};
use k8s_openapi::api::batch::v1::{CronJob, Job};
use k8s_openapi::api::core::v1::{ConfigMap, Pod, PodSpec};
use k8s_openapi::serde::de::DeserializeOwned;
use k8s_openapi::{Metadata, NamespaceResourceScope, Resource};
use kube::api::{DeleteParams, ObjectMeta};
use kube::config::KubeConfigOptions;
use kube::{
    api::{Api, ListParams},
    Client,
};
use kube::{Config, ResourceExt};
use log::{debug, info};
use regex::Regex;
use tokio::join;

const EXEMPTIONS: [&str; 1] = ["kube-root-ca.crt"];

#[derive(Debug, Args)]
pub struct CommandArgs {
    #[arg(short, long)]
    namespace: Option<String>,

    #[arg(short, long)]
    context: Option<String>,

    /// The kind of resource to clean up.
    #[arg(value_enum)]
    resource: Resources,

    /// Show more detailed logs (repeat to show more)
    #[arg(short, action=ArgAction::Count)]
    pub verbosity: u8,

    /// Do not perform any actions against the cluster.
    #[arg(long)]
    dry_run: bool,

    /// Delete resources that matches a regex.
    #[arg(short, long)]
    filter: Option<String>,

    /// Transforms the filter to a blacklist.
    #[arg(long)]
    inverse_filter: bool,
}

#[derive(ValueEnum, Debug, Clone)]
pub enum Resources {
    ConfigMap,
}

#[tokio::main()]
pub async fn handle(args: CommandArgs) -> Result<(), Box<dyn std::error::Error>> {
    if args.namespace.is_none() {
        debug!("No namespace specified, will use what's in the current context.");
    }

    let config = Config::from_kubeconfig(&KubeConfigOptions {
        context: args.context,
        cluster: None,
        user: None,
    })
    .await?;

    let client = Client::try_from(config)?;
    match args.resource {
        Resources::ConfigMap => {
            clean_config_maps(
                client,
                args.namespace.as_ref(),
                args.dry_run,
                args.filter,
                args.inverse_filter,
            )
            .await
        }
    }
}

async fn clean_config_maps(
    client: Client,
    namespace: Option<&String>,
    dry_run: bool,
    filter: Option<String>,
    inverse_filter: bool,
) -> Result<(), Box<dyn std::error::Error>> {
    let (
        config_maps,
        free_pods,
        deployments,
        replicasets,
        statefulsets,
        daemonsets,
        cronjobs,
        free_jobs,
    ) = join!(
        get_resources::<ConfigMap>(&client, namespace),
        get_ownerless_resources::<Pod>(&client, namespace),
        get_resources::<Deployment>(&client, namespace),
        get_resources::<ReplicaSet>(&client, namespace),
        get_resources::<StatefulSet>(&client, namespace),
        get_resources::<DaemonSet>(&client, namespace),
        get_resources::<CronJob>(&client, namespace),
        get_ownerless_resources::<Job>(&client, namespace),
    );

    let mut used_config_maps = get_config_map_references(free_pods?);
    used_config_maps.extend(get_config_map_references(deployments?).into_iter());
    used_config_maps.extend(get_config_map_references(replicasets?).into_iter());
    used_config_maps.extend(get_config_map_references(statefulsets?).into_iter());
    used_config_maps.extend(get_config_map_references(daemonsets?).into_iter());
    used_config_maps.extend(get_config_map_references(cronjobs?).into_iter());
    used_config_maps.extend(get_config_map_references(free_jobs?).into_iter());
    debug!("Done fetching resources from the Kubernetes API server.");

    let config_maps: HashSet<String> = config_maps?
        .into_iter()
        .map(|config_map| config_map.name_any())
        .collect();
    let unused_config_maps: HashSet<String> = config_maps
        .difference(&used_config_maps)
        .filter(|config_map| {
            let is_exempted = EXEMPTIONS.contains(&config_map.as_str());
            if is_exempted {
                debug!("Will not deleted {config_map} because it's exempted.")
            }
            !is_exempted
        })
        .filter(|config_map| {
            filter.as_ref().map_or(true, |regexp| {
                let is_matching = Regex::new(&regexp.as_str())
                    .unwrap()
                    .find(config_map.as_str())
                    .is_some();
                if inverse_filter {
                    if is_matching {
                        info!("Will not delete config map {config_map} because it's filtered-out.");
                        false
                    } else {
                        true
                    }
                } else {
                    if is_matching {
                        true
                    } else {
                        info!("Will not delete config map {config_map} because it's filtered-out.");
                        false
                    }
                }
            })
        })
        .cloned()
        .collect();
    info!(
        "There are {} config maps, {} are used, {} will be removed.",
        config_maps.len(),
        used_config_maps.len(),
        unused_config_maps.len()
    );
    unused_config_maps
        .iter()
        .for_each(|config_map| println!("{config_map}"));
    if dry_run {
        info!("Not deleting anything")
    } else {
        let _ = delete_resources::<ConfigMap>(
            &client,
            namespace,
            unused_config_maps.into_iter().collect(),
        )
        .await;
        info!("Unused config maps deleted.")
    }
    Ok(())
}

async fn get_resources<
    K: Resource<Scope = NamespaceResourceScope>
        + Metadata<Ty = ObjectMeta>
        + DeserializeOwned
        + Clone
        + Debug,
>(
    client: &Client,
    namespace: Option<&String>,
) -> Result<Vec<K>, Box<dyn std::error::Error>> {
    let client = Cow::Borrowed(client);
    let resources: Api<K> = match namespace {
        Some(ns) => Api::namespaced(client.into_owned(), ns),
        None => Api::default_namespaced(client.into_owned()),
    };
    let objects = resources.list(&ListParams::default()).await?;
    if !objects.items.is_empty() {
        debug!(
            "Got {} {}{} from the namespace {}",
            objects.items.len(),
            K::KIND,
            if objects.items.len() > 1 { "s" } else { "" },
            objects.items[0].metadata().namespace.as_ref().unwrap()
        );
    }
    Ok(objects.items)
}

async fn get_ownerless_resources<
    K: Resource<Scope = NamespaceResourceScope>
        + Metadata<Ty = ObjectMeta>
        + DeserializeOwned
        + Clone
        + Debug,
>(
    client: &Client,
    namespace: Option<&String>,
) -> Result<Vec<K>, Box<dyn std::error::Error>> {
    let result: Vec<K> = get_resources::<K>(client, namespace)
        .await?
        .into_iter()
        .filter(|resource| resource.metadata().owner_references.is_none())
        .inspect(|resource| {
            debug!(
                "Found {} {} without an owner.",
                K::KIND,
                resource.metadata().name.as_ref().unwrap(),
            )
        })
        .collect();
    Ok(result)
}

async fn delete_resources<
    K: Resource<Scope = NamespaceResourceScope>
        + Metadata<Ty = ObjectMeta>
        + DeserializeOwned
        + Clone
        + Debug,
>(
    client: &Client,
    namespace: Option<&String>,
    targets: Vec<String>,
) -> Result<(), Box<dyn std::error::Error>> {
    let client = Cow::Borrowed(client);
    let resources: Api<K> = match namespace {
        Some(ns) => Api::namespaced(client.into_owned(), ns),
        None => Api::default_namespaced(client.into_owned()),
    };
    for ref target in targets {
        resources
            .delete(target.as_str(), &DeleteParams::default())
            .await?;
    }
    Ok(())
}

trait HasPodSpec {
    fn pod_spec(&self) -> &PodSpec;
}

impl HasPodSpec for Pod {
    fn pod_spec(&self) -> &PodSpec {
        self.spec.as_ref().unwrap()
    }
}

impl HasPodSpec for Deployment {
    fn pod_spec(&self) -> &PodSpec {
        self.spec.as_ref().unwrap().template.spec.as_ref().unwrap()
    }
}

impl HasPodSpec for ReplicaSet {
    fn pod_spec(&self) -> &PodSpec {
        self.spec
            .as_ref()
            .unwrap()
            .template
            .as_ref()
            .unwrap()
            .spec
            .as_ref()
            .unwrap()
    }
}

impl HasPodSpec for StatefulSet {
    fn pod_spec(&self) -> &PodSpec {
        self.spec.as_ref().unwrap().template.spec.as_ref().unwrap()
    }
}
impl HasPodSpec for DaemonSet {
    fn pod_spec(&self) -> &PodSpec {
        self.spec.as_ref().unwrap().template.spec.as_ref().unwrap()
    }
}

impl HasPodSpec for CronJob {
    fn pod_spec(&self) -> &PodSpec {
        self.spec
            .as_ref()
            .unwrap()
            .job_template
            .spec
            .as_ref()
            .unwrap()
            .template
            .spec
            .as_ref()
            .unwrap()
    }
}

impl HasPodSpec for Job {
    fn pod_spec(&self) -> &PodSpec {
        self.spec.as_ref().unwrap().template.spec.as_ref().unwrap()
    }
}

fn get_config_map_references<
    K: Resource<Scope = NamespaceResourceScope>
        + Metadata<Ty = ObjectMeta>
        + DeserializeOwned
        + Clone
        + Debug
        + HasPodSpec,
>(
    resources: Vec<K>,
) -> HashSet<String> {
    resources
        .into_iter()
        .flat_map(|resource| extract_config_maps_from(&resource, resource.pod_spec()))
        .collect()
}

fn extract_config_maps_from<
    R: Resource<Scope = NamespaceResourceScope> + Metadata<Ty = ObjectMeta>,
>(
    resource: &R,
    pod_spec: &PodSpec,
) -> HashSet<String> {
    extract_config_maps_from_env_vars(resource, Cow::Borrowed(pod_spec))
        .into_iter()
        .chain(extract_config_maps_from_env_from(
            resource,
            Cow::Borrowed(pod_spec),
        ))
        .chain(extract_config_maps_from_volumes(
            resource,
            Cow::Borrowed(pod_spec),
        ))
        .chain(extract_config_maps_from_projected_volumes(
            resource,
            Cow::Borrowed(pod_spec),
        ))
        .collect()
}

fn extract_config_maps_from_volumes<
    R: Resource<Scope = NamespaceResourceScope> + Metadata<Ty = ObjectMeta> + ResourceExt,
>(
    owner: &R,
    pod_spec: Cow<PodSpec>,
) -> HashSet<String> {
    pod_spec.into_owned().volumes.map_or(HashSet::new(), |volumes| {
        volumes
            .into_iter()
            .filter_map(|volume| Some((volume.name, volume.config_map?.name?)))
            .inspect(|(volume, config_map)| {
                debug!(
                    "Reference to config map {config_map} found in volume {volume} in {} {} in namespace {}.",
                    R::KIND,
                    owner.name_any(),
                    owner.namespace().unwrap_or_default()
                )
            })
            .map(|(_, config_map)| config_map)
            .collect()
    })
}

fn extract_config_maps_from_projected_volumes<
    R: Resource<Scope = NamespaceResourceScope> + Metadata<Ty = ObjectMeta> + ResourceExt,
>(
    owner: &R,
    pod_spec: Cow<PodSpec>,
) -> HashSet<String> {
    pod_spec.into_owned().volumes.map_or(HashSet::new(), |volumes| {
        volumes
            .into_iter()
            .filter_map(|volume| Some((volume.name, volume.projected?.sources?)))
            .flat_map(|(volume, projections)| {
                projections
                    .into_iter()
                    .filter_map(move |projection| Some((volume.to_owned(), projection.config_map?.name?)))
            })
            .inspect(|(volume, config_map)| {
                debug!(
                    "Reference to config map {config_map} found in a projected volume {volume} in {} {} in namespace {}.",
                    R::KIND,
                    owner.name_any(),
                    owner.namespace().unwrap_or_default()
                )
            })
            .map(|(_, config_map)| config_map)
            .collect()
    })
}

fn extract_config_maps_from_env_vars<
    R: Resource<Scope = NamespaceResourceScope> + Metadata<Ty = ObjectMeta> + ResourceExt,
>(
    owner: &R,
    pod_spec: Cow<PodSpec>,
) -> HashSet<String> {
    pod_spec
        .into_owned()
        .containers
        .into_iter()
        .filter_map(|container| Some((container.name, container.env?)))
        .flat_map(|(container, env_vars)| env_vars.into_iter().map(move |env_var| (container.to_owned(), env_var)))
        .filter_map(|(container, env_var)| Some((container, env_var.name, env_var.value_from?.config_map_key_ref?.name?)))
        .inspect(|(container, env_var, config_map)|{
            debug!(
                "Reference to config_map {config_map} found in the env var {env_var} of container {container} in the pod spec of {} {} in namespace {}.",
                R::KIND,
                owner.name_any(),
                owner.namespace().unwrap_or_default()
            )
        })
        .map(|(_, _, config_map)| config_map)
        .collect()
}

fn extract_config_maps_from_env_from<
    R: Resource<Scope = NamespaceResourceScope> + Metadata<Ty = ObjectMeta>,
>(
    owner: &R,
    pod_spec: Cow<PodSpec>,
) -> HashSet<String> {
    pod_spec
        .into_owned()
        .containers
        .into_iter()
        .filter_map(|container| Some((container.name, container.env_from?)))
        .flat_map(|(container, sources)| sources.into_iter().map(move |source| (container.to_owned(), source)))
        .filter_map(|(container, source)| Some((container, source.config_map_ref?.name?)))
        .inspect(|(container, config_map)|{
            debug!(
                "Reference to config_map {config_map} found in the envFrom of container {container} in the pod spec of {} {} in namespace {}.",
                R::KIND,
                owner.name_any(),
                owner.namespace().unwrap_or_default()
            )
        })
        .map(|(_, config_map)| config_map)
        .collect()
}
