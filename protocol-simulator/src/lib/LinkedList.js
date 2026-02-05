// Implementation taken from:
// https://www.geeksforgeeks.org/javascript/implementation-linkedlist-javascript/
/** @typedef {import('$lib/types.js').Message} Message */


export class Node{
    /**
     * @param {Message} value
     */
    constructor(value)
    {
        this.value=value
        this.next=null
    }
}
export class LinkedList{
    constructor()
    {
        this.head=null
        this.length = 0
    }
    /**
     * @param {Message} value
     */
    append(value)
    {
        let newnode=new Node(value)
        this.length++
        if(!this.head)
        {
            this.head=newnode
            return
        }
        let current=this.head
        while(current.next)
        {
            current=current.next
        }
        current.next=newnode

    }
    /**
     * @returns {Message | null}
     */
    pop()
    {
        if (this.head != null){
            this.length--
            let val = this.head.value
            this.head = this.head.next
            return val
        }
        return null
    }
    printList(){
        let current=this.head
        let result=""
        while(current)
        {
            result+=current.value+'->'
            current=current.next
        }
        console.log(result+'null')
    }
}
