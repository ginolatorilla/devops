use clap::Subcommand;
pub mod clean;

#[derive(Subcommand, Debug)]
pub enum Commands {
    /// Cleans up unused Kubernetes resources.
    #[command()]
    Clean(clean::CommandArgs),
}
