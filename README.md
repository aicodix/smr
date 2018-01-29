
The results below are made using this tool on a: "Seagate Archive HDD ST8000AS0002"

Here the very noisy log visualized:

```
plot "64MiB.log" u ($3/1073741824):($4/1048576) w l t "64 MiB chunks", "128MiB.log" u ($3/1073741824):($4/1048576) w l t "128 MiB chunks", "256MiB.log" u ($3/1073741824):($4/1048576) w l t "256 MiB chunks", "512MiB.log" u ($3/1073741824):($4/1048576) w l t "512 MiB chunks", "1GiB.log" u ($3/1073741824):($4/1048576) w l t "1 GiB chunks"
```
![noisy log](noisy.png)

Using exponential moving average to smooth out a little:
(FYI: EMA filtering causes nonlinear phase delays.)

```
sum64MiB = 0
sum128MiB = 0
sum256MiB = 0
sum512MiB = 0
sum1GiB = 0
plot "64MiB.log" u ($3/1073741824):(sum64MiB=0.95*sum64MiB+0.05*$4, sum64MiB/1048576) w l t "64 MiB chunks", "128MiB.log" u ($3/1073741824):(sum128MiB=0.9*sum128MiB+0.1*$4, sum128MiB/1048576) w l t "128 MiB chunks", "256MiB.log" u ($3/1073741824):(sum256MiB=0.8*sum256MiB+0.2*$4, sum256MiB/1048576) w l t "256 MiB chunks", "512MiB.log" u ($3/1073741824):(sum512MiB=0.7*sum512MiB+0.3*$4, sum512MiB/1048576) w l t "512 MiB chunks", "1GiB.log" u ($3/1073741824):(sum1GiB=0.7*sum1GiB+0.3*$4, sum1GiB/1048576) w l t "1 GiB chunks"
```

![smoothed out](ema.png)

One thing can be seen immediately from the diagrams above:
The size of the persistent cache used for the device management of the HDD must be around 32GiB.

