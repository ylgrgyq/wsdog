# wsdog
 
A copycat for the great lib https://github.com/websockets/wscat but it does not need to install anything in advance like Node. Thus, it's suitable for environments where you cannot install software at will, such as your online servers.

## Installation

To install with [Homebrew](https://brew.sh/), run:

```
brew install ylgrgyq/homebrew-tap/wsdog
```

Alternatively, you can [download binary](https://github.com/ylgrgyq/wsdog/releases/latest) for your OS.

## Usage

```
Usage:
  wsdog [OPTIONS]

Application Options:
  -l, --listen=         listen on port
  -c, --connect=        connect to a WebSocket server
      --debug           enable debug log
      --no-color        Run without color
  -P, --show-ping-pong  print a notification when a ping or pong is received
  -s, --subprotocol=    optional subprotocol (default: )

Connect To A WebSocket Server Options:
  -o, --origin=         optional origin
  -x, --execute=        execute command after connecting
  -w, --wait=           wait given seconds after executing command (default: 2)
      --host=           optional host
  -n, --no-check        Do not check for unauthorized certificates
  -H, --header=         Set an HTTP header <header:value>. Repeat to set multiple like -H header1:value1 -H header2:value2.
      --auth=           Add basic HTTP authentication header <username:password>.
      --slash           Enable slash commands for control frames (/ping, /pong, /close [code [, reason]], /binary [Base64])

Help Options:
  -h, --help            Show this help message
```

## Example

```
$ wsdog -c ws://echo.websocket.org
Connected (press CTRL+C to quit)
> hi there
< hi there
> are you a happy parrot?
< are you a happy parrot?
```

For Binary Message, wsdog support write and print it in Base64 format. Such as

```
$ wsdog -c ws://echo.websocket.org --slash
Connected (press CTRL+C to quit)
> /binary SGVsbG8gd29ybGQh
<< SGVsbG8gd29ybGQh
```

Please note that the `--slash` option must be provided to active Slash Command Mode so we can use `/binary` command to send Binary Message in Base64. `SGVsbG8gd29ybGQh` is `Hello world!` in Base64. The leading `<<` means `wsdog` receives a Binary Message and print it's payload in Base64 format on the console. For Text Message, the payload will be print after `<` mark.

## License

MIT
