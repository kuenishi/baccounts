use serde::{Deserialize, Serialize};
use std::collections::{HashMap, HashSet};
use std::fs;
use std::path::PathBuf;

use anyhow::Context;
use log::{debug, error, info};
use url::Url;

#[derive(Serialize, Deserialize, Debug, Clone)]
#[allow(non_snake_case)]
pub struct Site {
    pub Url: String,
    pub Name: String,
    pub EncodedPass: String,
    pub Mail: String,
}

macro_rules! diff_string {
    ( $n:ident, $s:expr, $l:expr, $r:expr, $c:ident ) => {
        if $l.$n != $r.$n {
            if $s != "EncodedPass" {
                error!("{} mismatch at {}: {} != {}", $s, $l.Url, $l.$n, $r.$n);
            } else {
                error!(
                    "{} mismatch at {}: ({} bytes) != ({} bytes)",
                    $s,
                    $l.Url,
                    $l.$n.len(),
                    $r.$n.len()
                );
            }
            $c += 1;
        } else {
            debug!("{} ok", $s);
        }
    };
}

impl Site {
    pub fn update_pass(&mut self, pass: String) -> &mut Self {
        self.EncodedPass = pass;
        self
    }

    pub fn diff(&self, rhs: &Site, left: &PathBuf, right: &PathBuf) -> anyhow::Result<usize> {
        let mut count = 0;
        diff_string!(Url, "Url", self, rhs, count);
        diff_string!(Name, "Name", self, rhs, count);
        diff_string!(EncodedPass, "EncodedPass", self, rhs, count);
        diff_string!(Mail, "Mail", self, rhs, count);
        if count > 0 {
            println!("File\t| {:?} \t| {:?} \t|", left.display(), right.display());
            println!("Url \t| {} \t| {} \t|", self.Url, rhs.Url);
            println!("Name \t| {} \t\t| {} \t\t|", self.Name, rhs.Name);
            println!(
                "Pass \t| ({} bytes) \t\t| ({} bytes) \t\t|",
                self.EncodedPass.len(),
                rhs.EncodedPass.len()
            );
            println!("Mail \t| {} \t| {} \t|", self.Mail, rhs.Mail);
        }
        Ok(count)
    }
}

#[derive(Serialize, Deserialize, Debug, Clone)]
#[allow(non_snake_case)]
pub struct Profile {
    pub Name: String,
    pub Sites: HashMap<String, Site>,
    pub Default: bool,
}

impl Profile {
    pub fn new(name: &String) -> Self {
        Profile {
            Name: name.clone(),
            Sites: HashMap::new(),
            Default: false,
        }
    }

    pub fn list(&self) {
        for (key, value) in &self.Sites {
            println!("\t{}: \t{}", key, value.Mail);
        }
    }

    pub fn find_site(&self, site_name: &String) -> Option<&Site> {
        let mut found = 0;
        for (_name, site) in self.Sites.iter() {
            if site.Url.contains(site_name) {
                println!("Match found: {}", site.Url);
                found += 1;
            }
        }
        if found == 1 {
            for (_name, site) in self.Sites.iter() {
                if site.Url.contains(site_name) {
                    return Some(site);
                }
            }
        }
        error!("{} sites found. Only one should be found.", found);
        return None;
    }

    pub fn update_site(&mut self, site: Site) {
        let url = Url::parse(site.Url.as_str()).expect("Unable to parse URL");
        self.Sites.insert(url.host().unwrap().to_string(), site);
    }

    pub fn diff(&self, rhs: &Profile, left: &PathBuf, right: &PathBuf) -> anyhow::Result<usize> {
        let mut count = 0;
        let mut site_names = HashSet::new();
        for k in self.Sites.keys() {
            site_names.insert(k.clone());
        }
        for k in rhs.Sites.keys() {
            site_names.insert(k.clone());
        }
        for name in site_names {
            debug!("Comparing site {}: ", name);
            match (self.Sites.get(&name), rhs.Sites.get(&name)) {
                (Some(l), Some(r)) => {
                    // TODO: check difference
                    info!("Site {} found in both side", name);
                    count += l.diff(&r, &left, &right)?;
                }
                (Some(_), None) => {
                    error!("Site {} not found in {}", name, right.display());
                    count += 1;
                }
                (None, Some(_)) => {
                    error!("Site {} not found in {}", name, left.display());
                    count += 1;
                }
                (None, None) => {
                    anyhow::bail!("heh :P");
                }
            };
        }
        Ok(count)
    }
}

#[derive(Serialize, Deserialize, Debug)]
#[allow(non_snake_case)]
pub struct Baccounts {
    pub Profiles: Vec<Profile>,
    pub DefaultMail: String,
    pub Version: String,
}

impl Baccounts {
    pub fn new() -> Self {
        Baccounts {
            Profiles: Vec::new(),
            DefaultMail: String::new(),
            Version: String::new(),
        }
    }

    pub fn from_file(filename: &PathBuf) -> anyhow::Result<Self> {
        debug!("gpg --decrypt {}", filename.display());

        let path = filename.as_path();
        anyhow::ensure!(
            path.exists(),
            format!("The encrypted pass file {} does not exist.", filename.display())
        );

        let cmd_output = std::process::Command::new("gpg").arg("--decrypt").arg(filename).output()?;

        Ok(serde_json::from_slice(&cmd_output.stdout)?)
    }

    // Test purpose
    #[allow(dead_code)]
    pub fn from_raw_file(filename: &PathBuf) -> anyhow::Result<Self> {
        let data = fs::read_to_string(filename)?;
        let baccounts: Baccounts = serde_json::from_str(&data).context("Unable to parse file")?;
        Ok(baccounts)
    }

    pub fn list(&self) {
        for profile in &self.Profiles {
            println!("{}", profile.Name);
            profile.list();
        }
    }

    pub fn find_profile(&self, profile_name: &String) -> Option<&Profile> {
        for profile in &self.Profiles {
            if profile_name == "" {
                if profile.Default {
                    return Some(profile);
                }
            } else if profile.Name == *profile_name {
                return Some(profile);
            }
        }
        None
    }

    pub fn update_profile(&mut self, profile: Profile) -> Option<()> {
        //self.Profiles.push(profile);
        let mut index = 0;
        for p in &self.Profiles {
            if p.Name == profile.Name {
                self.Profiles[index] = profile;
                return Some(());
            }
            index += 1;
        }
        None
    }

    pub fn to_file(&self, name: &String, filename: &PathBuf) -> anyhow::Result<()> {
        //pub fn to_file(&self, name: &String, filename: &PathBuf) -> {
        let enc = std::process::Command::new("gpg")
            .arg("--encrypt")
            .arg("--armor")
            .arg("-r")
            .arg(name)
            .stdin(std::process::Stdio::piped())
            .stdout(std::process::Stdio::piped())
            .spawn()?;

        match enc.stdin.as_ref() {
            Some(stdin) => serde_json::to_writer_pretty(stdin, &self)?,
            None => anyhow::bail!("Failed to read stdin"),
        };
        let output = enc.wait_with_output().context("Unable to wait for gpg")?;

        Ok(fs::write(filename, &output.stdout)?)
    }

    // Test purpose
    #[allow(dead_code)]
    pub fn to_raw_file(&self, filename: &PathBuf) -> anyhow::Result<()> {
        let data = serde_json::to_string(&self)?;
        Ok(fs::write(filename, data)?)
    }

    pub fn diff(&self, rhs: &Baccounts, left: &PathBuf, right: &PathBuf) -> anyhow::Result<usize> {
        let mut count = 0;
        if self.Version != rhs.Version {
            error!("Version mismatch: {} != {}", self.Version, rhs.Version);
            count += 1;
        } else {
            debug!("Version ok: {}", self.Version);
        }
        if self.DefaultMail != rhs.DefaultMail {
            error!("DefaultMail mismatch: {} != {}", self.DefaultMail, rhs.DefaultMail);
            count += 1;
        } else {
            debug!("DefaultMail ok: {}", self.DefaultMail);
        }

        let mut profile_names = HashSet::new();
        for p in &self.Profiles {
            profile_names.insert(p.Name.clone());
        }
        for p in &rhs.Profiles {
            profile_names.insert(p.Name.clone());
        }

        for name in profile_names {
            println!("Comparing profile {}", name);

            match (self.find_profile(&name), rhs.find_profile(&name)) {
                (Some(l), Some(r)) => {
                    info!("Profile {} found in both side", name);
                    count += l.diff(&r, &left, &right)?;
                }
                (Some(_), None) => {
                    error!("Profile {} not found in {}", name, right.display());
                    count += 1;
                }
                (None, Some(_)) => {
                    error!("Profile {} not found in {}", name, left.display());
                    count += 1;
                }
                (None, None) => {
                    anyhow::bail!("heh :P");
                }
            };
        }
        Ok(count)
    }
}

#[cfg(test)]
mod tests {
    // Import the necessary modules and functions from your code
    use super::*;
    use std::path::PathBuf;
    //use crate::Baccounts;

    #[test]
    fn test_smoke() -> anyhow::Result<()> {
        // Write your test case here
        let b = Baccounts::new();
        b.to_raw_file(&PathBuf::from("/tmp/testfile.json"))?;
        let b2 = Baccounts::from_raw_file(&PathBuf::from("/tmp/testfile.json"))?;
        assert_eq!(b.Version, b2.Version);
        Ok(())
    }

    // Add more test cases as needed
}
