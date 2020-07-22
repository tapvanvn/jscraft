jscraft.require("template://module/base.js")

jscraft.build("base_template",{

    "content": ()=>{

        console.log("test content")
    }
})

var spawn_begin = Math.floor( (mapd.col - spawn_num) / 2 );
    
rowRun(i, (c)=>{ 
    if(spawn_begin > 0) {
        spawn_begin--;
    } else if ((spawn_num--) > 0){
        r=randomItem(); 
        c.i = design.items[r];
    }
})