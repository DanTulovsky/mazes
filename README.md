# mazes

## examples

```shell
go run client/client.go --op=create_solve \
  --random_path \
  --local_gui=false \
  -r 8 -c 8 -w 100 \
  --gen_draw_delay=0ms \
  --show_from_to_colors \
  --show_distance_colors \
  --solve_draw_delay=40ms \
  --create_algo=kruskal \
  --solve_algo=dijkstra
```

```shell
go run client/client.go --op=create_solve \
  --random_path \
  --local_gui=false \
  --show_from_to_colors \
  --show_distance_colors \
  --solve_draw_delay=20ms \
  --create_algo=ellers \
  --solve_algo=dijkstra \
  --gen_draw_delay=0ms \
  -r 80 -c 160 -w 8
```
