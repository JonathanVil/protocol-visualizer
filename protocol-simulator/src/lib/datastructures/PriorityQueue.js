import {Queue, Node} from '$lib/datastructures/Queue.js';

export class PriorityNode extends Node {
    constructor(value, priority) {
        super(value);
        this.priority = priority;
    }
}
// a priority queue sorted from lowest to highest
export class PriorityQueue extends Queue {
    constructor()
    {
        super();
    }
    // queues according to priority
    enqueue(value, priority){
        let newnode = new PriorityNode(value, priority)
        this.length++
        if(!this.head)
        {
            this.head=newnode
            return
        }
        let current=this.head
        while(current.priority <= priority) //while
        {
            if (current.next){
                current=current.next;
            } else {
                current.next=newnode; //add to end if lowest priority
            }
        }
        let next = current.next;
        current.next=newnode;
        newnode.next=next;
    }
    //pushes with priority 0
    push(value){
        this.enqueue(value, 0)
    }
}