# Craft js components to produce products

## Your code and file is classified to 3 category: 
1. Template: those file or resource will be used in multi projects and products as base component to build thing.
2. Layout: those file or resource specify, construct your final product. It belong to only one product.
3. Work: this is your final product those files is built by using (Template->Layout)

## The basic format of usage is:
- jscraft <template_dir> <layout_dir> <work_dir>
- The building process begins by read layout.json file in your layout directory in which each build step contain instruction to build one file or copy folder. See test/layout_1/layout.json 
```
    {
        "target":"work://script.js",
        "from":"layout://script.js"
    } 
```


## How to install?
- just run maintain.sh and reload your terminal.

## how to setup
1. 

## Basic struct and function