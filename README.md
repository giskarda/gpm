# gpm

Asynchronous web server in GO to return a list of file on the system.

Initially intended as a way to present RPM packages immediataly after upload without the need to a `create repo` sort of tool.

It has a lot of flow like:
- what if the packages has not yet been completely written
- no way to distinguish between release.

The nice part of this is to check the non and cached implementation of the same code.
