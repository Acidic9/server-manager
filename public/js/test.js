let myObj = { 0: 43, 5: 23 };
let obj = myObj[0];
for (let _ in myObj) {
    obj = parseInt(_);
    if (obj == 1) {
    }
}
