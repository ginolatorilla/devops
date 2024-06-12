use clap::Parser;
mod k8s;
use k8s::{kubeclean, Resources};

/// Cleans up unused Kubernetes resources.
#[derive(Parser, Debug)]
#[command(version, about, long_about = None)]
struct Args {
    #[arg(short, long)]
    namespace: Option<String>,

    /// The kind of resource to clean up.
    #[arg(value_enum)]
    resource: Resources,
}

fn main() {
    let mut clog = colog::default_builder();
    clog.filter(None, log::LevelFilter::Debug);
    clog.init();

    let args = Args::parse();
    match args.resource {
        Resources::ConfigMap => println!("I will clean all configmaps in {:?}.", args.namespace),
    }
    let _ = kubeclean(args.resource, args.namespace);
}
