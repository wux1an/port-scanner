# port-scanner
port scanner

## Usage

```ruby
$ port-scanner --help
A fast port scan tool based on full tcp connection

Usage:
  port-scanner [flags]

Examples:
  1. port-scanner -h 192.168.*.0 -p 80,443
  2. port-scanner -h 192.168.*.0 -p 80,443 -t 1 -n 50 -s -o result.txt

Flags:
  -b, --bar int         number of progress bar (default 5)
      --help            help for this command
  -h, --hosts string    hosts, eg: 10.0.0.1,10.0.0.5-10,192.168.1.*,192.168.10.0/24
      --no-color        disable colorful output
  -o, --output string   file path to output the opened ports
  -p, --ports string    ports, eg: 80 or 1-1024 or 1-1024,3389,8080
  -s, --shuffle         shuffle hosts and ports
  -n, --thread int      thread number (default 20)
  -t, --timeout int     timeout(second) of tcp connect (default 3)
  -v, --version         version

$ port-scanner -h 192.168.2.1-5 -p 1-500 -n 100 -t 1 -s -o r.txt
...
=================================================

Total:  5 hosts  x  500 ports
Start:  2022-04-13 12:04:35  Cost: 1.6174ms
Finish: 2022-04-13 12:04:35

IP               Opened  Ports
192.168.2.1          3   53 80 443
192.168.2.4          3   135 139 445

=================================================
```
