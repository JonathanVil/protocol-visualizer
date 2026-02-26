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

    find(predicate)
    {
        let current=this.head
        while(current)
        {
            if(predicate(current.value))
            {
                return current.value
            }
            current=current.next
        }
        return null
    }
    remove(predicate){
        let current = this.head
        let prev = null;

        while(current) {
            if (predicate(current.value)) {
                //If the node to remove is the head, we need to update the head pointer
                if (prev === null) {
                    this.head = current.next
                } else {
                    //If not head, we need to update the previous node's next pointer
                    prev.next = current.next
                }
                this.length--
                return current.value
            } else {
                prev = current;
                current = current.next;
            }
        }
        return null;
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