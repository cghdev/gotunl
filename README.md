# gotunl

gotunl is a command line client for Pritunl written in Go.

**Note:** gotunl package has been moved to this project under pkg/ directory.

## Installation:

Using go mod, requires go>=1.13:

```
git clone https://github.com/cghdev/gotunl.git
cd gotunl
go install
```

Alternatively, you can download a pre-compiled version from https://github.com/cghdev/gotunl/releases/latest


##  Usage:

```bash
Pritunl command line client

Usage:
  -c <profile>	Connect to profile ID or Name
  -d <profile>	Disconnect profile or "all"
  -l 		        List connections
  -v 		        Show version
```


## Examples:

```bash
$ ./gotunl -l
+----+------------------------+--------------+
| ID |          Name          |    Status    |
|----+------------------------+--------------|
|  1 | US VPN                 | Disconnected |
|  2 | UK VPN                 | Disconnected |
|  3 | Netherlands VPN        | Disconnected |
|  4 | Hong Kong VPN          | Disconnected |
|  5 | Test VPN               | Disconnected |
+----+------------------------+--------------+
$ ./gotunl -c 3
$ ./gotunl -c "Test VPN"
Enter the username: user1
Enter the password: *************
$ ./gotunl -l
+----+------------------------+--------------+---------------+---------------+---------------+
| ID |          Name          |    Status    | Connected for |   Client IP   |   Server IP   |
+----+------------------------+--------------+---------------+---------------+---------------+
|  1 | US VPN                 | Disconnected |               |               |               |
|  2 | UK VPN                 | Disconnected |               |               |               |
|  3 | Netherlands VPN        | Connected    | 16 secs       | 10.10.1.5     | 172.16.25.1   |
|  4 | Hong Kong VPN          | Disconnected |               |               |               |
|  5 | Test VPN               | Connected    | 8 secs        | 192.168.65.3  | 172.16.32.1   |
+----+------------------------+--------------+---------------+---------------+---------------+
$ ./gotunl -d all
```

