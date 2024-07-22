use clap::{Parser, Subcommand};
mod kube_clean;

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
    #[command(name = "kube-clean")]
    Kubeclean(kube_clean::CommandArgs),
}

fn main() {
    let args = Args::parse();

    match args.command {
        Commands::Kubeclean(args) => {
            configure_logger(args.verbosity);
            let _ = kube_clean::handle(args);
        }
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
