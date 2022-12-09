## boltdb exporter
The cli exports and dumps bolt db data as json/yaml format.

## Install
```
$ go get -u github.com/konoui/boltdb-exporter
```

## Usage
```
$ boltdb-exporter --db <database filename> --format yaml [--bucket <root bucket name> ...]
```

## Example
```
$ boltdb-exporter --db agent.db
{
  "metadata": {
    "agent-version": "1.44.2",
    "availability-zone": "ap-northeast-1a",
    "cluster-name": "default",
  (snip)
```
