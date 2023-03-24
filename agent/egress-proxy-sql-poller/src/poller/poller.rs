use crate::storage::storage::Storage;
use std::error::Error;

pub struct Poller;

impl Poller {
    pub fn exec<T: Storage + ?Sized>(g: Box<T>) -> Result<(), Box<dyn Error>> {
        let a = g.fetch_batch("123".to_string());
        println!("{}", a);
        Ok(())
    }
}
