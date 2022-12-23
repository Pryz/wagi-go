use std::env;

fn main() {
    let mut who: String = String::from("world");

    let args: Vec<String> = env::args().collect();
    for arg in args {
        let vec: Vec<&str> = arg.split("=").collect();
        if vec.len() == 2 && vec[0] == "who" {
            who = vec[1].to_string();
        }
    }

    println!("Content-Type: text/plain");
    println!("Status: 200");
    println!(); // Empty line between header and body
    println!("Hello {}", who);
}
