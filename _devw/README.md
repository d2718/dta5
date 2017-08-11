# Directory `_devw`

This directory contains the test world used for development (`dw2/`) and the help files (`help/`, obvs, also as they are being developed, also obvs).

### Directory structure for any world

```
main_directory/
    conf
    main.json
    descs/
    pc_dir/
    saves/
```

Commentary:

  * `main_directory/` (`dw2/`, in this case)
    This is the directory called as the command-line option when running the game. It contains all of the world-specific data.
    
  * `conf`
    Configuration file for the game. This has nothing to do with the world, but rather governs some parameters of how the game server runs. (For now, see the game's main file, `dta5/dta5.go` for details.)
  
  * `main.json`
    This is the world-building file that the game loader reads first. In any world of any reasonable size, it will mostly contain links to other files. For now, see the `dta5/load` package for the syntax of this file.
  
  * `descs/`
    The `dta5/desc` package will scan the files in this directory at load time for the descriptions of `Room`s and `Thing`s with specific descriptive text. A quick glance at the package and one of the files in this directory should make the syntax pretty clear.
  
  * `pc_dir/`
    Where `PlayerChar`s are saved when they log out.
  
  * `saves/`
    Where files containing the state of a currently running game are saved when the `save` command is issued to the game server. These files have a syntax that can be read by the `dta5/load` package, but aren't really meant for human consumption.
