# TODO

Until github issues

## UI

- [x] Use [chartjs](https://www.chartjs.org/docs/latest/getting-started/installation.html) to draw charts
- [ ] Fill the area between min_min and min_max, max_min and max_max, use chartjs [filling modes](https://www.chartjs.org/docs/latest/charts/area.html#filling-modes)
min_max, max_min `line` data should be hidden (`data.*.datasets.showLine: false`)
- [ ] Use smooth lines (e.g.: `datasets.*.tension: 0.4`)

## Backend

- [ ] Add "raw_reponses" to all sources, persist raw responses from the origin resources.
- [ ] Make accessable raw_reponses on HTTP
- [ ] Start to write tests
- [ ] Create min_max and max_min `line` dataset

## Bug

- [ ] Find the reason for incorrect min and max values (e.g.: Trier 16-23.09.2022)
