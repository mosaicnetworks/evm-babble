pragma solidity 0.4.8;

contract Test {
   
   uint localI = 1;
   
   event LocalChange(uint);

   function test(uint i) constant returns (uint){
        return i * 10;
   }

   function testAsync(uint i) {
        localI += i;
        LocalChange(localI);
   }
}
