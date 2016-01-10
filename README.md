CATS
----

![cats](./cats.jpg)

New Bartnet.

Development
-----------

Setup your postgresses.

```
createdb cats_test
export POSTGRES_CONN=postgres://grepory@localhost/cats_test
make migrate
make
```
