use config::{Config, ConfigError, Environment, File};
use lazy_static::lazy_static;
use serde_derive::Deserialize;
use std::env;
use std::sync::RwLock;

#[derive(Debug, Deserialize)]
#[allow(unused)]
pub struct Database {
    pub driver: String,
    pub address: String,
    pub username: String,
    pub password: String,
}

#[derive(Debug, Deserialize)]
#[allow(unused)]
pub struct Poller {
    pub poll_interval: u64,
}

#[derive(Debug, Deserialize)]
#[allow(unused)]
pub struct Settings {
    pub database: Database,
    pub poller: Poller,
}

impl Settings {
    pub fn new() -> Result<Self, ConfigError> {
        let run_mode = env::var("RUN_MODE").unwrap_or_else(|_| "development".into());
        let cfg = Config::builder()
            .add_source(
                File::with_name(&format!("~/.streams-egress-proxy/{}", run_mode)).required(false),
            )
            .add_source(Environment::with_prefix("STREAMS").separator("_"))
            .set_default("poller.poll_interval", 15)?
            .set_default("database.driver", "postgresql")?
            .set_default("database.address", "localhost:5432")?
            .set_default("database.username", "postgres")?
            .set_default("database.password", "root")?
            .build()?;
        cfg.try_deserialize()
    }
}

lazy_static! {
    pub static ref SETTINGS: RwLock<Settings> = RwLock::new(Settings::new().unwrap());
}
