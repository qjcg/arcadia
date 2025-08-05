# Diplomat

Generate PDF diplomas.

# Features
- cli tool takes as argument:
    - EITHER
        - instructor
        - course title
        - dates
        - recipient names
        - template
        - output directory
    - OR
        - config file

- template file
    - backround image w/ position
    - text nodes w/ positions

# Usage

Passing arguments via command line flags:

```
diplomat \
    -i 'Jerry Grapes' \
    -c 'Doing groovy stuff with TempleOS' \
    -d 'May 5-7 2015 (22.5h)' \
    -s 'Jack Groover, Jerry Funkytown, Janine Frankfurter'
```

Reading parameters from config file:

```
diplomat -f config.json
```

The contents of a template file:

```json
{
    "background": {
        "path": "/path/to/image.png",
        "x": 100,
        "y": 200
    },
    "text": {
        "instructor": {"x": 100, "y": 100},
        "dates": {"x": 100, "y": 100},
        "recipient": {"x": 100, "y": 100},
    }
}
```

ALTERNATE ideas for template file:
    - SVG
    - YAML
    - go (map or other struct?)
