# Daggerverse

### Requirements to develop
- dagger v0.12.0
- go 1.22+

### Develop an existing dagger module

```shell
cd <module-name>

# This will install the internal dagger packages, that are ignored in `.gitignore`
dagger develop
```

### Create a new dagger module

```shell
# Create a folder for a new module
mkdir <module-name>
cd <module-name>

# Create a module in the source ".", it will follow the monorepo structure for the .gitignore.
dagger init --name=<module-name> --sdk=go --source=.
```