use std::path::PathBuf;

use clap::{Parser, Subcommand};
use env_logger;

use anyhow::Context;
use log::{debug, error, info};
use xdg;

use rand::distributions::DistString;

use arboard::Clipboard;
#[cfg(target_os = "linux")]
use arboard::SetExtLinux;

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
    /// The name of profile
    #[clap(short = 'p', long = "profile", required = false, default_value = "")]
    profile: String,
}

#[derive(Debug, Subcommand)]
enum SubCommands {
    /// Test keys for encrypt and decrypt ready
    Test,

    #[clap(arg_required_else_help = true)]
    /// Add a new profile
    AddProfile {
        #[clap(long = "name", required = true, ignore_case = true)]
        name: String,
    },

    #[clap(arg_required_else_help = true)]
    /// Show the password (copyt to the clipboard)
    Show {
        #[clap(short = 's', long = "site", required = true, ignore_case = true)]
        site: String,
    },

    /// List all the sites in the profile (without password)
    List,

    #[clap(arg_required_else_help = true)]
    /// Generate password and save it to the profile
    Generate {
        #[clap(
            short = 'l',
            long = "len",
            required = false,
            default_value = "512",
            value_parser = clap::value_parser!(u64),
        )]
        len: u64,
        #[clap(long = "mail", required = true, ignore_case = true)]
        mail: String,
        // num_only: bool
        #[clap(long = "url", required = true)]
        url: String,
    },

    #[clap(arg_required_else_help = true, about = "Update password and save it to the profile")]
    Update {
        #[clap(
            short = 'l',
            long = "len",
            required = false,
            default_value = "83",
            value_parser = clap::value_parser!(u64),
        )]
        len: u64,
        //#[clap(long = "mail", required = true, ignore_case = true)]
        //mail: String,
        // num_only: bool
        #[clap(long = "site", required = true)]
        site: String,
    },
    Diff {
        /// Source file to compare (default: .config/baccounts..)
        #[clap(long)]
        lhs: Option<String>,

        /// Target file to compare (typically backup file)
        #[clap(long, required = true)]
        rhs: String,
    },
}

mod baccounts;
use baccounts::Baccounts;
use baccounts::Profile;

fn test() -> anyhow::Result<()> {
    info!("Testing the encryption & decryption environment...");

    let mut b = Baccounts::new();
    b.Version = "dummy bersion".to_string();
    b.DefaultMail = "dummy mail".to_string();
    let testfile = &std::path::PathBuf::from("/tmp/baccounts-test.json.asc");
    let profile_name = "kuenishi";
    b.to_file(&profile_name.to_string(), testfile)?;

    let b2: Baccounts = match std::process::Command::new("gpg").arg("--decrypt").arg(testfile).output() {
        Ok(cmd_output) => serde_json::from_slice(&cmd_output.stdout).expect("Unable to parse file"),
        Err(e) => {
            error!("Can't decrypt file {}: {}", profile_name, e);
            std::process::exit(1);
        }
    };

    assert_eq!(b.Version, b2.Version);
    assert_eq!(b.DefaultMail, b2.DefaultMail);
    Ok(())
}

fn generate_pass(len: u64) -> String {
    info!("Generating password with length {}", len);
    let printable_chars = rand::distributions::Alphanumeric;
    let mut rng = rand::thread_rng();

    let pass: String = printable_chars.sample_string(&mut rng, len as usize);
    debug!("Password generated: {} chars", pass.chars().count());
    pass
}

fn send2clipboard(pass: &String) -> anyhow::Result<()> {
    debug!("Password is being copied to the clipboard until the process ends");
    /*
          While the pass is in the clipboard, the user can paste it.
          This process will stay in memory until the clipboard is
          updated. See:
          https://docs.rs/arboard/3.3.0/arboard/trait.SetExtLinux.html
    */
    println!("Password should now be in clipboard. Waiting for termination...");
    let _ = Clipboard::new().unwrap().set().wait().text(pass.clone());
    Ok(())
}

fn main() -> anyhow::Result<()> {
    env_logger::init();
    info!("Baccounts: 💋 Password Manager");

    let confd = xdg::BaseDirectories::with_prefix("baccounts");
    let datafile = confd
        .get_config_file("baccounts.json.asc")
        .context("encoded file not found")?;
    let cli = Cli::parse();

    info!("Using profile '{}' (or default for empty)", cli.profile);

    match cli.subcommand {
        SubCommands::Test => test(),
        SubCommands::AddProfile { name } => {
            debug!("Adding profile: {}", name);
            let p = Profile::new(&name);
            let b = Baccounts::from_file(&datafile)?;
            match b.find_profile(&name) {
                Some(_) => anyhow::bail!("Profile already exists: {}", name),
                None => {
                    // TODO: Make sure this is notfound
                    info!("Adding new profile: {:?}", p);
                    unimplemented!();
                    //b.add_profile(p);
                    //b.to_file(&name, &datafile);
                    //info!("New profile saved to {}", datafile.display());
                }
            }
        }
        SubCommands::Show { site } => {
            debug!("Showing site: {}", site);
            let b = Baccounts::from_file(&datafile)?;
            let p = b.find_profile(&cli.profile).context("profile not found")?;
            debug!("Profile found: {}", p.Name);

            let Some(s) = p.find_site(&site) else {
                anyhow::bail!("Site not found: {}", site);
            };
            info!(
                "Password for {} ({}, {}) is being copied to the clipboard",
                s.Url, s.Name, s.Mail
            );
            send2clipboard(&s.EncodedPass)
        }
        SubCommands::List => {
            debug!("Listing sites");
            let b = Baccounts::from_file(&datafile)?;
            b.list();
            Ok(())
        }
        SubCommands::Generate { len, mail, url } => {
            info!("Generating password for site {} with user {}, length={}", url, mail, len);
            let mut b = Baccounts::from_file(&datafile)?;
            let p = b.find_profile(&cli.profile).context("profile not found")?;
            let profile_name = p.Name.clone();
            debug!("Profile found: {}", p.Name);

            if let Some(s) = p.find_site(&url) {
                error!("Site found ({}). Cannot generate password.", s.Name);
                anyhow::bail!("site found");
            }

            let name = url::Url::parse(&url).expect("Failed to parse --url");
            let pass = generate_pass(len);
            let site = baccounts::Site {
                Url: url,
                Name: name.host_str().unwrap().to_string(),
                EncodedPass: pass,
                Mail: mail,
            };
            let mut p2 = p.clone();
            p2.update_site(site);

            b.update_profile(p2).expect("Failed updating password");

            let tmpfile = confd.get_config_file("tmp-baccounts.json.asc").context("tempfile")?;
            b.to_file(&profile_name, &tmpfile)?;
            Ok(std::fs::rename(tmpfile, datafile.clone())?)
        }
        SubCommands::Update { len, site } => {
            debug!("Updating password for site {} length={}", site, len);

            let mut b = Baccounts::from_file(&datafile)?;
            let p = b.find_profile(&cli.profile).context("profile not found")?;
            let profile_name = p.Name.clone();
            debug!("Profile found: {}", p.Name);

            let s = p.find_site(&site).context("site not found")?;
            info!("Site found. Updating password for {} ({}, {})", s.Url, s.Name, s.Mail);

            let pass = generate_pass(len);

            let mut s2 = s.clone();
            s2.update_pass(pass);
            let mut p2 = p.clone();
            p2.update_site(s2);
            b.update_profile(p2).context("Updating password ok")?;

            let tmpfile = confd.get_config_file("tmp-baccounts.json.asc").context("tmpfile")?;
            b.to_file(&profile_name, &tmpfile)?;
            Ok(std::fs::rename(tmpfile, datafile.clone())?)
            //send2clipboard(&pass);
        }
        SubCommands::Diff { lhs, rhs } => {
            let lhs = match lhs {
                Some(f) => PathBuf::from(f),
                None => datafile,
            };
            let rhs = PathBuf::from(rhs.clone());
            let b0 = Baccounts::from_file(&lhs)?;
            let b1 = Baccounts::from_file(&rhs)?;
            info!("read: {} vs {}", lhs.display(), rhs.display());
            let _count = b0.diff(&b1, &lhs, &rhs);
            Ok(())
        }
    }
}
