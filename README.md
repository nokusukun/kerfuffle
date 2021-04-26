# Kerfuffle

Semi-full stack automated deployment pipeline.

Zero dependency successor to `nokusukun/Bandaid`.

## Requirements
* Go 1.16

## Install
```bash
$ git clone https://github.com/nokusukun/kerfuffle
$ cd kerfuffle
$ go build ./cmd/kerfuffle
-- OR --
$ go run ./cmd/kerfuffle
```
* If there are errors about missing packages then do `go mod tidy` first.

## Usage
Kerfuffle runs on port 80 for the public facing side and on port 8080 for the console.
The console lets you manage your applications.

## `.kerfuffle` files
`.kerfuffle` files are toml configuration files that lets you orchestrate the provision of the applications.
They compromise of three tags `provision`, `proxy` and `cloudflare`. `meta` is reserved for future use.


Example: https://github.com/nokusukun/odi-chat/blob/master/.kerfuffle
```toml
[meta]
    name = "Odi Chat"
    
[provision.init]
    run = [
        ["yarn"]
    ]

[provision.client]
    run = [
        ["yarn", "run", "start"]
    ]
    envs = []
    base_dir = ""

[proxy.client]
    host = ["chat.noku.pw"]


[cloudflare.client]
    host = ["chat.noku.pw"]
    zone = "noku.pw"
    proxied = true
```

### `provision` tag
Tags are always accompanied by an identifier. These allow kerfuffle to identify between daemons in your application.

Note: `[provision.init]` is special as it's always executed first before everyone else.

###`provision` fields
* `run`
    * a 2d string array compromised of the command and arguments
* `envs`
    * an array of environment variable pairs to be passed to the executables.
* `base_dir`
    * the path as to where the commands will be executed

### `proxy` tag
The proxy tags contains the data to allow kerfuffle to route the traffic between the installed applications.

**NOTE**: In order for this to work, kerfuffle will always pass an `APP_PORT` environment variable in which
the application has to bind to. See example for clarity.

###`proxy` fields
* `host`
    * an array of addresses to which the app binds to.

### `cloudflare` tag
If a cloudflare tag exists, then Kerfuffle will automatically configure cloudflare provided that the 
token key is present on the cf-zones folder (usually stored on `./cf-zones`)

###`cloudflare` fields
* `host`
    * an array of addresses to which kerfuffle will create an A record pointing to the machine's IP address.
* `zone`
    * the name of the zone where the record will be stored.
* `proxied`
    * enables all of the cloudflare features. Turn off if you just want to have cloudflare act as a DNS server.
    
### Examples
* https://github.com/nokusukun/odi-chat
* https://github.com/nokusukun/sample-express