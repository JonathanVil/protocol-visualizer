// Implementation taken from:
// https://www.geeksforgeeks.org/javascript/implementation-linkedlist-javascript/

export class Node{
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
    }
    append(value)
    {
        let newnode=new Node(value)
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
    pop()
    {
        if (this.head != null){
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
