// Implementation taken from:
// https://www.geeksforgeeks.org/javascript/implementation-linkedlist-javascript/


export class Node{
    /**
     * @param {any} value
     */
    constructor(value)
    {
        this.value=value
        this.next=null
    }
}
export class Queue {
    constructor()
    {
        this.head=null
        this.length = 0
    }
    /**
     * @param {any} value
     */
    push(value)
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
     * @returns {any}
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
