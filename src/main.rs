use clap::{ArgAction, Parser, Subcommand, ValueEnum};
mod k8s;
use k8s::kubeclean;
use k8s_openapi::{api::core::v1::ConfigMap, Resource};

fn main() {
    let args = Args::parse();

    match args.command {
        Commands::Kubeclean {
            namespace,
            resource,
            verbose,
            dry_run,
            filter,
            inverse_filter,
        } => {
            let mut clog = colog::default_builder();
            let log_level = match verbose {
                0 => log::LevelFilter::Warn,
                1 => log::LevelFilter::Info,
                2 => log::LevelFilter::Debug,
                _ => log::LevelFilter::Trace,
            };

            clog.filter(None, log_level);
            clog.init();

            let resource_kind = match resource {
                Resources::ConfigMap => ConfigMap::KIND,
            };
            let _ = kubeclean(resource_kind, namespace, dry_run, filter, inverse_filter);
        }
    }
}

/// Gino's DevOps tools
#[derive(Parser, Debug)]
#[command(version, about, long_about = None)]
struct Args {
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    /// Cleans up unused Kubernetes resources.
    Kubeclean {
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
    },
}

#[derive(ValueEnum, Debug, Clone)]
pub enum Resources {
    ConfigMap,
}
