use clap::{ArgAction, Parser, ValueEnum};
mod k8s;
use k8s::kubeclean;
use k8s_openapi::{api::core::v1::ConfigMap, Resource};

fn main() {
    let args = Args::parse();

    let mut clog = colog::default_builder();
    let log_level = match args.verbose {
        0 => log::LevelFilter::Warn,
        1 => log::LevelFilter::Info,
        2 => log::LevelFilter::Debug,
        _ => log::LevelFilter::Trace,
    };

    clog.filter(None, log_level);
    clog.init();

    let resource_kind = match args.resource {
        Resources::ConfigMap => ConfigMap::KIND,
    };
    let _ = kubeclean(
        resource_kind,
        args.namespace,
        args.dry_run,
        args.filter,
        args.inverse_filter,
    );
}

/// Cleans up unused Kubernetes resources.
#[derive(Parser, Debug)]
#[command(version, about, long_about = None)]
struct Args {
    #[arg(short, long)]
    namespace: Option<String>,

    /// The kind of resource to clean up.
    #[arg(value_enum)]
    resource: Resources,

    /// Show more detailed logs (repeat to show more)
    #[arg(short, action=ArgAction::Count)]
    verbose: u8,

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
