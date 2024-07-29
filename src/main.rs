use clap::{Parser, Subcommand};
mod kube_clean;

/// Gino's DevOps tools
#[derive(Parser, Debug)]
#[command(version, about, long_about = None)]
struct Args {
    #[command(subcommand)]
    group: Groups,
}

#[derive(Subcommand, Debug)]
enum Groups {
    Kubernetes {
        #[command(subcommand)]
        command: KubernetesCommands,
    },
}

#[derive(Subcommand, Debug)]
enum KubernetesCommands {
    /// Cleans up unused Kubernetes resources.
    #[command()]
    Clean(kube_clean::CommandArgs),
}

fn main() {
    let args = Args::parse();

    match args.group {
        Groups::Kubernetes { command } => match command {
            KubernetesCommands::Clean(args) => {
                configure_logger(args.verbosity);
                let _ = kube_clean::handle(args);
            }
        },
    }
}

fn configure_logger(verbosity: u8) {
    let mut clog = colog::default_builder();
    let log_level = match verbosity {
        0 => log::LevelFilter::Warn,
        1 => log::LevelFilter::Info,
        2 => log::LevelFilter::Debug,
        _ => log::LevelFilter::Trace,
    };

    clog.filter(None, log_level);
    clog.init();
}
