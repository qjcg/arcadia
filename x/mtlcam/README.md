# mtlcam

Pull down the latest Montreal traffic camera images using the
[open GeoJSON data](https://donnees.ville.montreal.qc.ca/dataset/cameras-observation-routiere)
provided by the City of Montreal.


## Usage

```sh
$ mtlcam -h

mtlcam: Download Montreal traffic camera images
Data source: https://donnees.ville.montreal.qc.ca/dataset/cameras-observation-routiere

  -c int
        max concurrent downloads (default 20)
  -d    print debug messages
  -p string
        parent directory for downloaded files (default "images")
  -v    print version

$ mtlcam
[...]

$ ls -R images
[...]
```
