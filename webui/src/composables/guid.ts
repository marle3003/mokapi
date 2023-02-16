export function useGuid() {
    function createGuid(): string {
        function _p8(s: boolean = false) {  
            var p = (Math.random().toString(16)+"000000000").substring(2,8);  
            return s ? "-" + p.substring(0,4) + "-" + p.substring(4,4) : p ;  
         }  
         return 'id' + _p8() + _p8(true) + _p8(true) + _p8();  
    }

    return {createGuid}
}