use clap::{Parser, Subcommand, ValueEnum};
use env_logger;
#[macro_use]
extern crate log;
use xdg;

use pgp::{
    composed::message::Message, composed::signed_key::*, crypto::sym::SymmetricKeyAlgorithm,
    Deserializable,
};

use rand::rngs::StdRng;
use rand::SeedableRng;

use std::{fs, io::Cursor, io::Read};

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
    #[clap(
        short = 'p',
        long = "profile",
        required = false,
        //value_name = "",
        default_value = "",
    )]
    profile: String,
}

#[derive(Debug, Subcommand)]
enum SubCommands {
    // Test keys for encrypt and decrypt ready
    Test,

    #[clap(arg_required_else_help = true)]
    Show {
        #[clap(short = 's', long = "site", required = true, ignore_case = true)]
        site: String,
    },
    List,

    #[clap(arg_required_else_help = true)]
    Generate {
        #[clap(
            short = 'l',
            long = "len",
            required = false,
            default_value = "8",
            value_parser = clap::value_parser!(u64),
        )]
        len: u64,
        #[clap(long = "mail", required = true, ignore_case = true)]
        mail: String,
        // num_only: bool
        #[clap(long = "url", required = true)]
        url: String,
    },

    #[clap(arg_required_else_help = true)]
    Update {
        #[clap(
            short = 'l',
            long = "len",
            required = false,
            default_value = "8",
            value_parser = clap::value_parser!(u64),
        )]
        len: u64,
        #[clap(long = "mail", required = true, ignore_case = true)]
        mail: String,
        // num_only: bool
        #[clap(long = "url", required = true)]
        url: String,
        #[clap(long = "new", required = true)]
        new: String,
    },
}

#[derive(Debug, Clone, ValueEnum)]
enum Format {
    Csv,
    Json,
}

fn main() {
    env_logger::init();

    println!("Hello, world!");
    debug!("hello");

    let confd = xdg::BaseDirectories::with_prefix("baccounts").unwrap();
    let pkey_file = confd.get_config_file("kuenishi-public-key.key");
    debug!("{}", pkey_file.display());

    // let pubkey = fs::read(pkey_file).unwrap();
    let mut f = fs::File::open(&pkey_file).unwrap();
    //.context("Trying to load pkey fron config")?;
    let pk = SignedPublicKey::from_bytes(f).unwrap();

    let msg = Message::new_literal("none", "hogehoge");

    let mut rng = StdRng::from_entropy();
    let new_msg = msg
        .encrypt_to_keys(&mut rng, SymmetricKeyAlgorithm::AES128, &[&pk])
        .unwrap();
    print!("{}", new_msg.to_armored_string(None).unwrap());

    let cli = Cli::parse();

    info!("Using profile '{}'", cli.profile);

    match cli.subcommand {
        SubCommands::Test => {}
        SubCommands::Show { site } => {
            info!("Showing site: {}", site);
            unimplemented!();
        }
        SubCommands::List => {
            unimplemented!();
        }
        SubCommands::Generate { len, mail, url } => {
            info!(
                "Generating password for site {} with user {}, length={}",
                url, mail, len
            );
            unimplemented!();
        }
        SubCommands::Update {
            len,
            mail,
            url,
            new,
        } => {
            info!(
                "Updating password for site {} with user {}, length={}",
                url, mail, len
            );
            unimplemented!();
        }
    }
}
