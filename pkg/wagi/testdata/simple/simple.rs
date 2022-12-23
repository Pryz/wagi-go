// In a WAGI module one of the following header must be set:
//    `context-type: MEDIA_TYPE`: `text/plain` or `application/javascript`
//    `location: FULL_URL`: not sure what this is used for ?
fn main() {
    println!("Content-Type: text/plain");
    println!("Status: 200");
    println!(); // Empty line between header and body
    println!("hello world");
}
