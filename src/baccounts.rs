use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::fs;
use std::io::Write;
use std::path::PathBuf;

use url::Url;

#[derive(Serialize, Deserialize, Debug, Clone)]
#[allow(non_snake_case)]
pub struct Site {
    pub Url: String,
    pub Name: String,
    pub EncodedPass: String,
    pub Mail: String,
}

impl Site {
    pub fn update_pass(&mut self, pass: String) -> &mut Self {
        self.EncodedPass = pass;
        self
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

    pub fn from_file(filename: &PathBuf) -> Self {
        debug!("gpg --decrypt {}", filename.display());
        if !filename.as_path().try_exists().unwrap() {
            error!(
                "The encrypted pass file {} does not exist.",
                filename.display()
            );
            std::process::exit(1);
        }
        match std::process::Command::new("gpg")
            .arg("--decrypt")
            .arg(filename)
            .output()
        {
            Ok(cmd_output) => {
                let baccounts: Baccounts =
                    serde_json::from_slice(&cmd_output.stdout).expect("Unable to parse file");
                baccounts
            }
            Err(e) => {
                error!("Can't decrypt file {}: {}", filename.display(), e);
                std::process::exit(1);
            }
        }
    }

    // Test purpose
    #[allow(dead_code)]
    pub fn from_raw_file(filename: &PathBuf) -> Self {
        let file = match fs::File::open(filename) {
            Ok(f) => f,
            Err(e) => {
                error!("Can't open file {}: {}", filename.display(), e);
                std::process::exit(1);
            }
        };
        let baccounts: Baccounts = serde_json::from_reader(&file).expect("Unable to parse file");
        baccounts
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

    pub fn to_file(&self, name: &String, filename: &PathBuf) {
        let enc = std::process::Command::new("gpg")
            .arg("--encrypt")
            .arg("--armor")
            .arg("-r")
            .arg(name)
            .stdin(std::process::Stdio::piped())
            .stdout(std::process::Stdio::piped())
            .spawn()
            .expect("Unable to launch gpg");

        serde_json::to_writer_pretty(enc.stdin.as_ref().unwrap(), &self)
            .expect("Unable to write file");
        let output = enc.wait_with_output().expect("Unable to wait for gpg");

        let mut file = match fs::File::create(filename) {
            Ok(f) => f,
            Err(e) => {
                error!("Can't open file {}: {}", filename.display(), e);
                std::process::exit(1);
            }
        };
        file.write_all(&output.stdout)
            .expect("Unable to write file");
    }

    // Test purpose
    #[allow(dead_code)]
    pub fn to_raw_file(&self, filename: &PathBuf) {
        let file = match fs::File::create(filename) {
            Ok(f) => f,
            Err(e) => {
                error!("Can't open file {}: {}", filename.display(), e);
                std::process::exit(1);
            }
        };
        serde_json::to_writer_pretty(file, &self).expect("Unable to write file");
    }
}

#[cfg(test)]
mod tests {
    // Import the necessary modules and functions from your code
    use super::*;
    use std::path::PathBuf;
    //use crate::Baccounts;

    #[test]
    fn test_smoke() {
        // Write your test case here
        let b = Baccounts::new();
        b.to_raw_file(&PathBuf::from("/tmp/testfile.json"));
        let b2 = Baccounts::from_raw_file(&PathBuf::from("/tmp/testfile.json"));
        assert_eq!(b.Version, b2.Version);
    }

    // Add more test cases as needed
}
