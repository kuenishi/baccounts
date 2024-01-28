use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::fs;
use std::io::prelude::*;
use std::path::PathBuf;

#[derive(Serialize, Deserialize, Debug)]
#[allow(non_snake_case)]
pub struct Site {
    pub Url: String,
    pub Name: String,
    pub EncodedPass: String,
    pub Mail: String,
}

#[derive(Serialize, Deserialize, Debug)]
#[allow(non_snake_case)]
pub struct Profile {
    pub Name: String,
    pub Sites: HashMap<String, Site>,
    pub Default: bool,
}

impl Profile {
    fn new(name: &String) -> Self {
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
        for (name, site) in self.Sites.iter() {
            if site.Url.contains(site_name) {
                return Some(site);
            }
        }
        return None;
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
    fn new() -> Self {
        Baccounts {
            Profiles: Vec::new(),
            DefaultMail: String::new(),
            Version: String::new(),
        }
    }

    pub fn from_file(filename: &PathBuf) -> Self {
        let mut file = fs::File::open(filename).expect("Unable to open file");
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
}
