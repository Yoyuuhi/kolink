# kolink
Easily output simple call graph for golang between every two packages.

# Features
Based on abstract structure tree (AST):
- Find (caller) files which used the types defined in callee package
- Find (caller) files which used the functions defined in callee package
- Find (caller) files which used the methods defined in callee package

And output .png graphs by each callee file using graphviz.

*The analysis is only based on AST between caller and callee packages. (The variables retrived by function/method from other package can not be solved)

# Synopsis
## Settings
`kolink` needs configuration file ( `kolink.yml` ) to running.

`kolink.yml` example is the following.

```yaml
repositoryName: kolink

requestDefs:
  - outDir: kolinkGraph
    callee:
      dir: graph
      ignoreNewFunc: true
      ignoreTest: true
    caller:
      dir: ./
      ignoreTest: true
    fileParams:
      - fileName: graph.go
        layout: circo
        split: true
        attributes:
          - name: ranksep
            value: 2
  - outDir: kolinkGraph
    callee:
      dir: request
      ignoreNewFunc: true
      ignoreTest: true
    caller:
      dir: ./
      ignoreTest: true
```
- `repositoryName`: name of your repository
- `requestDefs`: define caller and callee for call graph
  - `outDir`: directory for output .png files
  - `callee`: package info for called type, function
    - `dir`: directory of callee package
    - `ignoreNewFunc`: ignore Newhoge functions if true
    - `ignoreTest`: ignore hogetest files if true
  - `caller`: package info for caller
    - `dir`: directory of caller package
    - `ignoreTest`: ignore hogetest files if true
  - `fileParames`: setting params for output
    - `fileName`: target callee file for setting
    - `layout`: graphviz layout
    - `split`: output graph by type/func if true (default output by callee file)
    - `attributes`: can add graphviz attributes

## Example
![image](https://github.com/Yoyuuhi/kolink/assets/53810181/7117e6ba-0231-4031-a3b4-85cebcf4611e)

# License
MIT
