use clap::{ArgAction, Parser};
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

    /// Log level to show.
    #[arg(short, action=ArgAction::Count)]
    verbose: u8,
}

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

    match args.resource {
        Resources::ConfigMap => println!("I will clean all configmaps in {:?}.", args.namespace),
    }
    let _ = kubeclean(args.resource, args.namespace);
}
