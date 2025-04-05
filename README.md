# Mikrograf

This tool is an input "plugin" for Telegraf to collect data from Mikrotik devices

## Configuration

This tool supposed to be running as exec plugin in telegraf. Whole configuration is in one env variabla **MIKROGRAF_TARGET_HOSTS**
which consists of the router configration lines divided by semicolon (;).

Configuration line is:
```
SCHEMA://[username:password]@HOST:PORT?ignoreCertificate=true&ignoreComments=comment1,comment2&modules=module1,module2&ignoreDisabled=false
```
...where:
* SCHEMA = http/https.
* username,password = as usual.
* HOST:PORT = as usual.
* ignoreCertificate = when set to true will ignore not trusted TLS certificate. **False** by default.
* ignoreComments = comma divided list of comments that if assigned to objects will make that objects excluded. Empty by default.
* modules = comma divided list of modules to be used. **system_resources** will be used by default.
* ignoreDisabled = when set to false will not ignore disabled objects. **True** by default.

Current modules are:
* interface
* interface_wireguard_peers
* interface_wireless_registration
* ip_dhcp_server_lease
* ip_firewall_connection
* ip_firewall_filter
* ip_firewall_mangle
* ip_firewall_nat
* ipv6_firewall_connection
* ipv6_firewall_filter
* ipv6_firewall_mangle
* ipv6_firewall_nat
* system_resourses
* system_script
* all - will include all modules

... so this one is the perfectly fine configuration:
```
http://admin:password@192.168.88.1:5987?modules=interface,interface_wireguard_peers,interface_wireless_registration,ip_dhcp_server_lease,ipv6_firewall_filter,ipv6_firewall_nat,ipv6_firewall_mangle,system_script,system_resourses&ignoreComments=ignore_this,ignore_that;https://admin22:password@192.168.99.1?ignoreCertificate=true&modules=all&ignoreComments=ignore_what&ignoreDisabled=false
```

You can use this plugin like this in Telegraf:
```toml
[[inputs.exec]]
  commands = ["/usr/local/bin/mikrograf"]
  environment = ["MIKROGRAF_TARGET_HOSTS=http://admin:password@192.168.88.1:5987?modules=interface,interface_wireguard_peers,interface_wireless_registration,ip_dhcp_server_lease,ipv6_firewall_filter,ipv6_firewall_nat,ipv6_firewall_mangle,system_script,system_resourses&ignoreComments=ignore_this,ignore_that;https://admin22:password@192.168.99.1?ignoreCertificate=true&modules=all&ignoreComments=ignore_what"]

  # Optional
  # timeout = "5s"

  # Optional
  # name_suffix = ""

  # Optional. Better not to ignore
  # ignore_error = false

  # data format MUST be line protocol
  data_format = "lineprotocol"
```

... or for clarity of the configuration you can separate hosts between multiple plugins:
```toml
[[inputs.exec]]
  commands = ["/usr/local/bin/mikrograf"]
  environment = ["MIKROGRAF_TARGET_HOSTS=https://admin22:password@192.168.99.1?ignoreCertificate=true&modules=all&ignoreComments=very_secret_shit"]
  data_format = "lineprotocol"

[[inputs.exec]]
  commands = ["/usr/local/bin/mikrograf"]
  environment = ["MIKROGRAF_TARGET_HOSTS=http://admin:password@192.168.88.1:5987?modules=interface,interface_wireguard_peers,interface_wireless_registration,ip_dhcp_server_lease,ipv6_firewall_filter,ipv6_firewall_nat,ipv6_firewall_mangle,system_script,system_resourses&ignoreComments=ignore_this,ignore_that&ignoreDisabled=false"]
  data_format = "lineprotocol"
```
