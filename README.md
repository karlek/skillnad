WIP
---
This project is a *work in progress*. The implementation is *incomplete* and
subject to change. The documentation can be inaccurate.

Skillnad
=========

Skillnad is a collection of various programs to glitch data for recreational
purposes. Skillnad means difference in Swedish and the name is a reference to my
favorite short movie of all time, ["Everything is Poetry Baby"](https://www.youtube.com/watch?v=BE0BY9tORhQ).

Installation
------------

`go get github.com/karlek/skillnad/cmd/skillnad`

Generate an image
-----------------

```shell
$ skillnad -x 0 -y 1 manifest.png
```

Flags:
------

* __x:__
	Amount of pixel-sorting on the x-axis.
* __y:__
	Amount of pixel-sorting on the y-axis.
* __xy:__
	Sort the x-axis before the y-axis.
* __yx:__
	Sort the y-axis before the x-axis.

Example
--------

This is the original image:

![Original](https://github.com/karlek/skillnad/blob/master/manifest.png?raw=true)

And this is the sorted image:

![Sorted](https://github.com/karlek/skillnad/blob/master/out.png?raw=true)

Produced with:

```shell
$ skillnad -x 0 -y 1 manifest.png
```

Public domain
-------------
I hereby release this code into the [public domain](https://creativecommons.org/publicdomain/zero/1.0/).
