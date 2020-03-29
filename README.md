# mempool-cli

[mempool.space](https://mempool.space/) for the terminal
![mempool screenshot](https://github.com/gchaincl/mempool/raw/master/share/screenshot.png)

Mempool is a mempool visualizer for the Bitcoin blockchain inspired by
[mempool.space](https://mempool.space/). It connects to the same API endpoint but the block rendering happens on your terminal.

:warning: This software is being under developement, things might break :warning:
# Install
Get a pre-built [release](https://github.com/gchaincl/mempool/releases/latest)

### development version
```bash
go get -u github.com/mempool/mempool-cli
```

### Docker
```bash
docker run -it mempool/mempool-cli
```

# Usage
### Key bindings
Key               | Description
------------------|--------------------------------------
<kbd>Ctrl+c</kbd> | Quit
<kbd>f</kbd>      | Opens Tx search
<kbd>click</kbd>  | Opens info for a selected block

# TODO
- [x] Transaction Tracking (by Tx ID) (using 'f' key)
- [x] Block details on click
- [ ] Graphs
- [ ] Tx weight per second progress bar
- [ ] Custom API endpoint 
