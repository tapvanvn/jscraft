var _base = "base"

function baseFunction() {
    console.log(_base);
}

jscraft.template("base_template", ()=>{

    var b = 1
    var c = 2 + b
    var a = c * c
    console.log(a)

    jscraft.fetch("content")
})