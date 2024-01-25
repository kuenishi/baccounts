use clap::{Parser, Subcommand, ValueEnum};

use env_logger;
#[macro_use] extern crate log;

#[derive(Debug, Parser)]
#[clap(
    name = env!("CARGO_PKG_NAME"),
    version = env!("CARGO_PKG_VERSION"),
    author = env!("CARGO_PKG_AUTHORS"),
    about = env!("CARGO_PKG_DESCRIPTION"),
    arg_required_else_help = true,
)]
struct Cli {
    #[clap(subcommand)]
    subcommand: SubCommands,
    /// server url
    #[clap(short = 's', long = "server", value_name = "URL", default_value = "localhost:3000")]
    server: String,
}


#[derive(Debug, Subcommand)]
enum SubCommands {
    // Test keys for encrypt and decrypt ready
    Test,
    #[clap(arg_required_else_help = true)]
    Get {
        /// log format
        #[clap(
            short = 'f',
            long = "format",
            required = true,
            ignore_case = true,
            value_enum,
        )]
        format: Format,
    },
    /// post logs, taking input from stdin
    Post,
}


#[derive(Debug, Clone, ValueEnum)]
enum Format {
    Csv,
    Json,
}

fn main() {
    pritty_env_logger::try_init();
    
    println!("Hello, world!");
    debug!("hello");

    let cli = Cli::parse();
    match cli.subcommand {
        SubCommands::Test => {
            
        },
        SubCommands::Get { format } => {
            match format {
                Format::Csv => unimplemented!(),
                Format::Json => unimplemented!(),
            }
        },
        SubCommands::Post => unimplemented!(),
    }
}
