# wsdog
 
A copycat for the great lib https://github.com/websockets/wscat but it does not need to install anything in advance like Node. Thus, it's suitable for environments where you cannot install software at will, such as your online servers.

## Installation



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
      --slash           Enable slash commands for control frames (/ping, /pong, /close [code [, reason]])

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

## License

MIT
