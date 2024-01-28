
use serde::{Serialize, Deserialize};
use std::fs;
use std::collections::HashMap;
use std::io::prelude::*;

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
	pub Name:   String,
	pub Sites:   HashMap<String, Site>,
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

    pub fn from_file(filename: &str) -> Self {
        let mut file = fs::File::open(filename).expect("Unable to open file");
        let mut contents = String::new();
        file.read_to_string(&mut contents).expect("Unable to read file");
        let baccounts: Baccounts = serde_json::from_str(&contents).expect("Unable to parse file");
        baccounts
    }

    pub fn list(&self) {
        for profile in &self.Profiles {
            println!("{}", profile.Name);
            profile.list();
        }
    }
}