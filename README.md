# prometheus-storagebox-exporter

This tool talks to the [Hetzner
API](https://robot.your-server.de/doc/webservice/de.html#storage-box) and
gets a list of all [Storage
Boxes](https://www.hetzner.de/storage/storage-box) in your account and exports their statistics as Prometheus metrics on port `<host>:9509/metrics`.

## Authentication
Sadly the old Hetzner API only accepts BasicAuth as an authenticaton method for their API so this exporter needs your customer number and password for your Hetzner account.
These variables gets passed to the tool as environment variables: `HETZNER_USER` and `HETZNER_PASS`

## Exported Metrics 
```
# HELP storagebox_disk_quota Total diskspace in MB
# TYPE storagebox_disk_quota gauge
storagebox_disk_quota{host="FSN1-BX1234",id="1234",location="FSN1",name="Backup",product="BX10",server="u12345.your-storagebox.de"} 102400
# HELP storagebox_disk_usage Total used diskspace in MB
# TYPE storagebox_disk_usage gauge
storagebox_disk_usage{host="FSN1-BX1234",id="1234",location="FSN1",name="Backup",product="BX10",server="u12345.your-storagebox.de"} 23256
# HELP storagebox_disk_usage_data Used diskspace by files in MB
# TYPE storagebox_disk_usage_data gauge
storagebox_disk_usage_data{host="FSN1-BX1234",id="1234",location="FSN1",name="Backup",product="BX10",server="u12345.your-storagebox.de"} 23256
# HELP storagebox_disk_usage_snapshots Used diskspace by snapshots in MB
# TYPE storagebox_disk_usage_snapshots gauge
storagebox_disk_usage_snapshots{host="FSN1-BX1234",id="1234",location="FSN1",name="Backup",product="BX10",server="u12345.your-storagebox.de"} 0
# HELP storagebox_location_hash Number representation of the location short name
# TYPE storagebox_location_hash gauge
storagebox_location_hash{host="FSN1-BX1234",id="1234",location="FSN1",name="Backup",product="BX10",server="u12345.your-storagebox.de"} 3.868487187e+09
# HELP storagebox_host_system_hash Number representation of the location short name
# TYPE storagebox_host_system_hash gauge
storagebox_host_system_hash{host="FSN1-BX1234",id="1234",location="FSN1",name="Backup",product="BX10",server="u12345.your-storagebox.de"} 1.813712476e+09

```
