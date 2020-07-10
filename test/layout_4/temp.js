jscraft.require("template://module/base.js")

jscraft.template("temp_1", ()=> {
    jscraft.build("base_template",{
        "content": ()=>{
            jscraft.fetch("temp_content")
        }
    })
})
