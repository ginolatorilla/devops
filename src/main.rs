use clap::{Parser, Subcommand};
mod kubernetes;

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
        command: kubernetes::Commands,
    },
}

fn main() {
    let args = Args::parse();

    match args.group {
        Groups::Kubernetes { command } => match command {
            kubernetes::Commands::Clean(args) => {
                configure_logger(args.verbosity);
                let _ = kubernetes::clean::handle(args);
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
