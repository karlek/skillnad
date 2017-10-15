Create JPEG artifacts
======================

```shell
$ convert -strip -gaussian-blur 0.05 -quality 1% input.png output.jpg
```

Make a glitched and broken JPEG avaiable for more glitching
============================================================

```shell
$ convert broken.jpg output.png
```

Sometimes convert changes the image, then try gimp.