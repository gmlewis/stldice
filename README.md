# stldice

Experimental code to work with STL and voxelized designs in Go

## Overview

Traditional CAD tools were designed to perform constuctive solid geomtery (CSG)
(aka "Boolean") operations on simple, regular shapes such as cylinders, cubes,
spheres, and toroids. However, as 3D printers are continually producing higher
resolution parts, the designs being printed are becoming more complex. The
existing tools provided by major CAD corporations simply are not capable of
using the traditional algorithms that got them this far, even when running
on distributed machines in large data centers. The algorithms themselves have
exhausted their usefulness.

It is time for a paradigm shift.

## Voxels

Voxels are the paradigm shift that is needed. Just as desktop (2D) printers
started out simple in the form of dot matrix printers and have improved
exponentially in quality, 3D printers are experiencing the same fantastic
progression. Likewise, the concept of rasterizing 2D images into pixels
translates into 3D by rasterizing designs into 3D pixels or "voxels".
Using this technique, a design can be rasterized at any desired resolution,
as high as the 3D printer supports.

The greatest advantage of voxels, though, is the ability to perform extremely
detailed and complex boolean operations that traditional mesh algorithms are
not able to do.

This repo provides a set of tools that enable high resolution
[STL](https://en.wikipedia.org/wiki/STL_(file_format))
designs to be diced, cut, then recontructed back to STL so that the large corpus
of tools available today can be used to further manipulate the meshes and so
that the designs can be printed on any 3D printers.

The suite of tools consists of:

* `binvox` - package to read/write binvox files
* `stl` - package that provides STL merge capabilities
* `stl2svx` - experimental Kubernetes cluster to batch process voxel designs
* `stldice` - dices up STL meshes into one or more
  [`vox`](https://raw.githubusercontent.com/ephtracy/voxel-model/master/MagicaVoxel-file-format-vox.txt)
  files
* `tri2stl` - combines `tri` files back into STL mesh files
* `voxcut-dice` - writes to stdout many `voxcut` commands to cover a full model
* `voxcut` - performs boolean operations on `binvox` files
* `vox2tri` - converts `vox` files to `tri` files
* `vshell` - start of experiment to represent a voxel model by its shell only

----------------------------------------------------------------------

Enjoy!

----------------------------------------------------------------------

# License

Copyright 2018 Glenn Lewis. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
