var VM = require("vm");

const sandbox = {
    x: 2,
    tes : function(a, b){
        return a + b;
    }
}

VM.createContext(sandbox);

const code = `
   var y = tes(x, 20);
`

VM.runInContext(code, sandbox);

console.log(sandbox);