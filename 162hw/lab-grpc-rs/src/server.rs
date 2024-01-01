//! The gRPC server.
//!

use crate::{log, rpc::kv_store::*, SERVER_ADDR};
use anyhow::Result;
use tonic::{transport::Server, Request, Response, Status};

use tokio::sync::RwLock;
use std::sync::Arc;
use std::collections::HashMap;


pub struct KvStore {
    map: Arc<RwLock<HashMap<Vec<u8>, Vec<u8>>>>
}

#[tonic::async_trait]
impl kv_store_server::KvStore for KvStore {
    async fn example(
        &self,
        req: Request<ExampleRequest>,
    ) -> Result<Response<ExampleReply>, Status> {
        log::info!("Received example request.");
        Ok(Response::new(ExampleReply {
            output: req.into_inner().input + 1,
        }))
    }

    // TODO: RPC implementation
    async fn echo(
        &self,
        req: Request<EchoRequest>,
    ) -> Result<Response<EchoReply>, Status> {
        log::info!("Received echo request.");
        Ok(Response::new(EchoReply {
            output: req.into_inner().msg,
        }))
    }

    async fn put(
        &self,
        req: Request<PutRequest>,
    ) -> Result<Response<PutReply>, Status> {
        log::info!("Received put request.");        
        let inner = req.into_inner();
        let key = inner.key;
        let value = inner.value;
        {
            let mut write_lock = self.map.write().await;
            write_lock.insert(key, value);
        }
        Ok(Response::new(PutReply{}))
    }

    async fn get(
        &self,
        req: Request<GetRequest>,
    ) -> Result<Response<GetReply>, Status> {
        log::info!("Received get request.");
        let key = req.into_inner().key;
        {
            let read_lock = self.map.read().await;
            match read_lock.get(&key) {
                Some(value) => Ok(Response::new(GetReply {
                    value: value.clone(),
                })),
                None => Err(tonic::Status::new(tonic::Code::NotFound, "Key does not exist.")),
            }
        }
    }
}

pub async fn start() -> Result<()> {
    let svc = kv_store_server::KvStoreServer::new(KvStore {map: Arc::new(RwLock::new(HashMap::new()))});

    log::info!("Starting KV store server.");
    Server::builder()
        .add_service(svc)
        .serve(SERVER_ADDR.parse().unwrap())
        .await?;
    Ok(())
}
