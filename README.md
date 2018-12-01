# gotunl

gotunl is a command line client for Pritunl written in Go.


Usage:

```bash
Pritunl command line client

Usage:
  -c <profile>	Connect to profile ID or Name
  -d <profile>	Disconnect profile or "all"
  -l 		        List connections
  -v 		        Show version
```

Examples:
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
+----+------------------------+--------------+
| ID |          Name          |    Status    |
|----+------------------------+--------------|
|  1 | US VPN                 | Disconnected |
|  2 | UK VPN                 | Disconnected |
|  3 | Netherlands VPN        | Connected    |
|  4 | Hong Kong VPN          | Disconnected |
|  5 | Test VPN               | Connected    |
+----+------------------------+--------------+
$ ./gotunl -d all
```

