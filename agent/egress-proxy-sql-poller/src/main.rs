use crate::poller::poller::Poller;
use crate::settings::settings::SETTINGS;
use crate::storage::storage::new_storage;
use tokio::io::{self};
use tokio::time::{self, Duration};

mod poller;
mod settings;
mod storage;

#[tokio::main]
async fn main() -> io::Result<()> {
    let interval_seconds = SETTINGS.read().unwrap().poller.poll_interval;
    let mut interval = time::interval(Duration::from_secs(interval_seconds));
    loop {
        interval.tick().await;
        tokio::spawn(async {
            let driver = SETTINGS.read().unwrap().database.driver.to_string();
            let store = new_storage(driver.as_str());
            Poller::exec(store).expect("streams.proxy.egress: event traffic routing task failed");
        });
    }
}
