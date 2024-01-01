const REQUEST_BUF_SIZE: usize = 1024;

use std::env;
use std::net::{Ipv4Addr, SocketAddrV4};

use crate::args;

use crate::http::*;
use crate::stats::*;

use clap::Parser;
use tokio::net::{TcpListener, TcpStream};

use tokio::fs::File;
use tokio::io::{AsyncReadExt, AsyncWriteExt};
use tokio::fs;
use std::path::Path;
use std::ops::Not;
use std::path::PathBuf;
use anyhow::Result;

pub fn main() -> Result<()> {
    // Configure logging
    // You can print logs (to stderr) using
    // `log::info!`, `log::warn!`, `log::error!`, etc.
    env_logger::Builder::new()
        .filter_level(log::LevelFilter::Info)
        .init();

    // Parse command line arguments
    let args = args::Args::parse();

    // Set the current working directory
    env::set_current_dir(&args.files)?;

    // Print some info for debugging
    log::info!("HTTP server initializing ---------");
    log::info!("Port:\t\t{}", args.port);
    log::info!("Num threads:\t{}", args.num_threads);
    log::info!("Directory:\t\t{}", &args.files);
    log::info!("----------------------------------");

    // Initialize a thread pool that starts running `listen`
    tokio::runtime::Builder::new_multi_thread()
        .enable_all()
        .worker_threads(args.num_threads)
        .build()?
        .block_on(listen(args.port))
}

async fn listen(port: u16) -> Result<()> {
    // Hint: you should call `handle_socket` in this function.
    //todo!("TODO: Part 2")
    let listener = TcpListener::bind(format!("0.0.0.0:{}", port)).await?;

    loop {
        let (socket, _) = listener.accept().await?;

        tokio::spawn(async move {
            // Process each socket concurrently.
            handle_socket(socket).await
        });
    }

    Ok(())
}

// Handles a single connection via `socket`.
async fn handle_socket(mut socket: TcpStream) -> Result<()> {
    //todo!("TODO: Part 3")
    let request = parse_request(&mut socket).await?;

    if request.method == "GET" {
        //build file
        let path_string = format!("./{}", request.path);
        let mut file_path = Path::new(&path_string);
        let index_str = format_index(&path_string);
        let mut full_path = PathBuf::from(&request.path);

        if (file_path).is_dir() {
            let index_path = Path::new(&index_str);
            if index_path.exists() {
                // if index.html exists
                file_path = index_path;
            } else {
                //directory listing, since index.html doesn't exist
                // let path_str = file_path.to_str().unwrap();
                // let dir_listing = format_index(path_str);
                // socket.write_all(dir_listing.as_bytes()).await?;
                start_response(&mut socket, 200).await?;
                send_header(&mut socket, "Content-Type", "text/html").await?;
                end_headers(&mut socket).await?;
                socket.write_all("<html><body><br/>".as_bytes()).await?;
                let mut dir_list = tokio::fs::read_dir(&file_path).await.unwrap();
                // for path in dir_list {
                //     let path_str = path.as_os_str().into_str().unwrap();
                //     socket.write_all(&path_str).await?;
                // }
                match full_path.parent() {
                    Some(parent) => {
                        let path_str = parent.to_str().unwrap();
                        let link_str = format_href(&path_str, "..");
                        socket.write_all(link_str.as_bytes()).await?;
                        socket.write_all("<br/>".as_bytes()).await?;
                    }
                    None => println!("No parent directory found."),
                }

                while let Some(entry) = dir_list.next_entry().await? {
                    let path_str = entry.file_name().into_string().unwrap();
                    //let full_path = format!("{}/{}", request.path, path_str);
                    // let mut full_path = PathBuf::from(&request.path);
                    let mut full_path_cpy = full_path.clone();
                    full_path_cpy.push(&path_str);
                    let full_path_str = full_path_cpy.into_os_string().into_string().unwrap();
                    let link_str = format_href(&full_path_str, &path_str);
                    socket.write_all(link_str.as_bytes()).await?;
                    socket.write_all("<br/>".as_bytes()).await?;
                }
                socket.write_all("</body></html>".as_bytes()).await?;
            }
        }



        //println!("{}",file_path.display());
        if (file_path).is_dir().not() {
            match File::open(&file_path).await {
                Ok(mut file) => {
                    let mime = get_mime_type(&(file_path.to_str().unwrap()));
    
                    let content_len = fs::metadata(&file_path).await?.len();
    
                    start_response(&mut socket, 200).await?;
                    send_header(&mut socket, "Content-Type", mime).await?;
                    send_header(&mut socket, "Content-Length", &content_len.to_string()).await?;
                    end_headers(&mut socket).await?;
    
                    let mut buffer = vec![0; REQUEST_BUF_SIZE];
    
                    loop {
                        match file.read(&mut buffer).await {
                            Ok(0) => break,
                            Ok(n) => {
                                //println!("{:?}", buffer);
                                socket.write_all(&buffer[0..n]).await?;
                            }
                            
                            Err(err) => {
                                log::warn!("Error when reading file: {}", err);
                                break;
                            }
                        }
                    }
                }
    
    
                Err(_) => {
                    //404 error
                    start_response(&mut socket, 404).await?;
                    end_headers(&mut socket).await?;
                }
            }
        }
        
    } else {
        start_response(&mut socket, 405).await?;
        send_header(&mut socket, "Content-Type", "text/plain").await?;
        end_headers(&mut socket).await?;
    }

    Ok(())

}

// You are free (and encouraged) to add other funtions to this file.
// You can also create your own modules as you see fit.
