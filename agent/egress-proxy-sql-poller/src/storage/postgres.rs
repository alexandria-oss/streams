use crate::storage::storage::Storage;

pub struct Postgres;

impl Storage for Postgres {
    fn fetch_batch(&self, _batch_id: String) -> String {
        return "getting batch with psql".to_string();
    }
}

impl Postgres {
    pub fn new() -> Postgres {
        Postgres {}
    }
}
