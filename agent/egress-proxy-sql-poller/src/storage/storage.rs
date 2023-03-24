use crate::storage::postgres::Postgres;
use std::string::ToString;

pub trait Storage {
    fn fetch_batch(&self, batch_id: String) -> String;
}

pub struct NoopStorage;

impl Storage for NoopStorage {
    fn fetch_batch(&self, _batch_id: String) -> String {
        "fetching batch from noop".to_string()
    }
}

pub fn new_storage(driver: &str) -> Box<dyn Storage> {
    match driver {
        "postgres" => Box::new(Postgres::new()),
        _ => Box::new(NoopStorage),
    }
}
