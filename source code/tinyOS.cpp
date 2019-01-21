#include "component.h"
//this part used to control the sequece when execute processes
//version 1: scheduler, static list,fixed length
//version 2: dispatcher, dynamic list


/*
void scheduler::loader(procptr pptr){ 
	wait.enQueue(pptr); 
} //load proc(pointer) to wait queue

int scheduler::swap(interpreter &IM, processor &PM){
	int i;
	int counter=0;
	int r=1;
	procptr outPause=0;

	while(wait.deQueue(run)!=-1){	//wait queue is not empty, means there is still pra need to run
		run->state=1;
		printf("D%-5d process ID:%d ", ID,run->ID);
		RES=70;
		systemChecker.check(RES, sysLog);
		counter=0;	//cycle of cycle is to 5
		while(counter< priorityCounterTable[run->priority] && r==1){
			r=IM.exer(PM,*run); //run
			counter++;
		}


		if(r==1){ //timp cycle arrives
			wait.enQueue(run);
			run->state=0;
			printf("D%-5d process ID:%d ", ID,run->ID);
			RES=71;
			systemChecker.check(RES, sysLog);
		}
		else
			if(r>=10){
				printf("D%-5d caller process ID:%d, system call NO:%d ",ID,run->ID,r-10); //syscall checked, at this time r-10 == sysfun ID
				RES=73;
				systemChecker.check(RES, sysLog);
				sysCall.copy(r-10,sysTem); //got system fun according to r
				for(i=0; i<run->MData.mSize; i++){
					sysTem->MData.write(i,run->MData.read(i)); //copy MData to syscall ??system will change waitqueue
				}
				sysTem->callerID=run->ID; //indicate who calls sysTem
				pause[r-10].enQueue(run);	//proc to wait sysfun completed //simplified!!! //have changed
				run->state=2;
				printf("D%-5d process ID:%d, pause cause:%d ",ID,run->ID,sysTem->ID);	//casue need be to changed
				RES=72;
				systemChecker.check(RES, sysLog);
				wait.enQueue(sysTem); //sysfun wait to fun
				sysTem->state=0;

				printf("D%-5d system function ID:%d, caller ID:%d ", ID,sysTem->ID,sysTem->callerID);
				RES=76;
				systemChecker.check(RES, sysLog);
			//	sysTem->~process();
				r=1;
			}
			else
				if(r==3){
					printf("D%-5d system function ID:%d, caller process ID:%d ", ID,run->ID,run->callerID);
					RES=74;
					systemChecker.check(RES, sysLog);
					pause[run->ID].deQueue(outPause);
					if(outPause->ID==run->callerID){ //compare ID
						for(i=0; i<run->MData.mSize; i++)
							outPause->MData.write(i,run->MData.read(i));
						wait.enQueue(outPause);
						outPause->state=0;

						printf("D%-5d process ID:%d, pause cause:%d ", ID,outPause->ID,run->ID);
						RES=75;
						systemChecker.check(RES, sysLog);
						r=1;	
					}

				}
				else
					if(r==0 || r==-1){
						if(r==0) run->state=3;
						else run->state=4;
						r=1;	//a process finished or interpreter over flow
					}
	}
	printf("D%-5d interpreter ID:%d, processor ID:%d ",ID, IM.ID, PM.ID);
	RES=79; //queue is empty is normal when scheduler finished
	systemChecker.check(RES, sysLog);
	return 0;
}

int scheduler::swap2(interpreter &IM, processor &PM){
	//int i;
	int counter=0;
	int r=1;
	procptr outPause=0;

	while(wait.deQueue(run)!=-1){	//wait queue is not empty, means there is still pra need to run
		run->state=1;
		printf("D%-5d process ID:%d ", ID,run->ID);
		RES=70;
		systemChecker.check(RES, sysLog);
		counter=0;	//cycle of cycle is to 5
		while(counter< priorityCounterTable[run->priority] && r==1){
			r=IM.exer(PM,*run); //run
			counter++;
		}
		if(r==1){ //timp cycle arrives
			run->state=0;
			wait.enQueue(run);
			printf("D%-5d process ID:%d ", ID,run->ID);
			RES=71;
			systemChecker.check(RES, sysLog);
		}
		else
			if(r>=10){
				printf("D%-5d caller process ID:%d, system call NO:%d ",ID,run->ID,r-10); //syscall checked, at this time r-10 == sysfun ID
				RES=73;
				systemChecker.check(RES, sysLog);
				sysCall.copy(r-10,sysTem); //got system fun according to r
				//for(i=0; i<run->MData.mSize; i++){
				//	sysTem->MData.write(i,run->MData.read(i)); //copy MData to syscall ??system will change waitqueue
				//}
				//not copy MData!!! use the way as follows:
				sysTem->MData.mem=run->MData.mem;	//sysTem->MData.mem pointer to the same unit as caller
				sysTem->MData.mSize=run->MData.mSize;	//make sysTem->MData same mSize as caller

				sysTem->callerID=run->ID; //indicate who calls sysTem
				pause[r-10].enQueue(run);	//proc to wait sysfun completed //simplified!!! //have changed
				run->state=2;
				printf("D%-5d process ID:%d, pause cause:%d ",ID,run->ID,sysTem->ID);	//casue need be to changed
				RES=72;
				systemChecker.check(RES, sysLog);
				wait.enQueue(sysTem); //sysfun wait to fun
				sysTem->state=0;
				printf("D%-5d system function ID:%d, caller ID:%d ", ID,sysTem->ID,sysTem->callerID);
				RES=76;
				systemChecker.check(RES, sysLog);
			//	sysTem->~process();
				r=1;
			}
			else
				if(r==3){
					run->state=3;
					printf("D%-5d system function ID:%d, caller process ID:%d ", ID,run->ID,run->callerID);
					RES=74;
					systemChecker.check(RES, sysLog);
					pause[run->ID].deQueue(outPause);
					if(outPause->ID==run->callerID){ //compare ID
					//	for(i=0; i<run->MData.mSize; i++)
					//		outPause->MData.write(i,run->MData.read(i));
						wait.enQueue(outPause);
						outPause->state=0;
										
						printf("D%-5d process ID:%d, pause cause:%d ", ID,outPause->ID,run->ID);
						RES=75;
						systemChecker.check(RES, sysLog);
						r=1;	

						run->MData.mem=0;	//destory the tem sysfun copy, set run->MData.mem=0 to avoid deconstruct caller's MData 
						delete run;

					}

				}
				else
					if(r==0 || r==-1){
						if(r==0){
							run->state=3;
							finished.enQueue(run);
							printf("D%-5d process ID:%d ",ID,run->ID);
							RES=77;
							systemChecker.check(RES, sysLog);
						}
						else{
							run->state=4;
							printf("D%-5d process ID:%d ",ID,run->ID);
							RES=78;
							systemChecker.check(RES, sysLog);
						}
						r=1;	//a process finished or interpreter over flow
					}
					else
						if(r==5 || r==6) //read system part
						{
							int address;
							int len;
							int i;
							
							if(run->priority!=0){	//only sysfun can visit this part
								printf("error!");
							}
							else{
								address=run->MRgst.read(1); //start address of system share to read or write
								len=run->MRgst.read(2);	//the lenght to read or write

								if(r==5){	//read system share to the lastest part of the MData 
									for(i=0; i<len; i++){ //eg. systemShare[0~2]={99,98,97},  then the last 3 unit of run->MData is 99,98,97
										run->MData.write(run->MData.mSize-len+i,systemShare[address+i]);
									}
									printf("D%-5d process ID:%d ", ID,run->ID);
									RES=100;
									systemChecker.check(RES, sysLog);
								}
								else{	//write lastest part of the MData into system share
									for(i=0; i<len; i++){
										systemShare[address+i]=run->MData.read(run->MData.mSize-len+i);
									}
									printf("D%-5d process ID:%d ", ID,run->ID);
									RES=101;
									systemChecker.check(RES, sysLog);
								}
								
								run->state=0;
								wait.enQueue(run);	//mov the process into wait queue
								printf("D%-5d process ID:%d ", ID,run->ID);
								RES=71;
								systemChecker.check(RES, sysLog);
							}
							r=1;	
						}
	}
	printf("D%-5d interpreter ID:%d, processor ID:%d ",ID, IM.ID, PM.ID);
	RES=79; //queue is empty is normal when scheduler finished
	systemChecker.check(RES, sysLog);
	return 0;
}
*/

//new scheduler
void dispatcher::loader(procptr pptr){ 
	//wait.enQueue(pptr);
	int i;
	i=pcbManagement();	//get a pcb from system pcb set
	if(i!=-1){
		pcbSet[i].link(*pptr);	//choise a pcb from pcb set and link it with the process
		readyb.enQueue(&pcbSet[i]);	//loader the pcb into ready queue
	}
	else{
	
		RES=184;
		if(systemChecker.showLevel(RES, sysLog)){
			printf("D%-5d process ID:%d ",ID,pptr->ID);
			systemChecker.check(RES, sysLog);
		}
	}

} //load proc(pointer) to wait queue

void dispatcher::loaderSysfun(procptr pptr){ 
	//wait.enQueue(pptr);
	int i;
	i=pcbManagement();	//get a pcb from system pcb set
	if(i!=-1){
		pcbSet[i].link(*pptr);	//choise a pcb from pcb set and link it with the process
		//readyb.enQueue(&pcbSet[i]);	//loader the pcb into ready queue	
		readyb.insertToHead(&pcbSet[i]);
	}
	else{
		RES=184;
		if(systemChecker.showLevel(RES, sysLog)){
			printf("D%-5d process ID:%d ",ID,pptr->ID);
			systemChecker.check(RES, sysLog);
		}
	}

} //load proc(pointer) to wait queue


int dispatcher::swap2(interpreter &IM){
	//int i;
	int counter=0;	//counter about running cycle times to swap processer
	int r=1;	//record the running condition after each cycle
	//procptr outPause=0;
	pcbptr outPauseb=0;	//point to the pcb out from wait queue

	while(readyb.deQueue(runb)!=-1){	//wait queue is not empty, means there is still pra need to run

		//tool_pcbStatistics();	//<<<<<<a tool to about pcb statistics>>>>>>

		//run->state=1;
		runb->state=1;	//process choosed to run

		RES=170;
		if(systemChecker.showLevel(RES, sysLog)){
			printf("D%-5d process ID:%d ", ID,runb->ID);				
			systemChecker.check(RES, sysLog);
		}

		counter=0;	//cycle of cycle is to 5
		
		while(counter< priorityCounterTable[runb->priority] && r==1){
			r=IM.exer(*(runb->pptr)); //run
			runb->exetime=runb->pptr->exetime;
			counter++;
		}
		
		if(runb->pptr->update==1){	//some property about the process has changed
			runb->update(runb->pptr);
			runb->pptr->update=0;
		}
		
		//*************************************
		//used block statement for each case
		//eg.
		//case 1:{
		//		statements;
		//		}
		//		break;	
		//*************************************
		switch(r){
			case -1:{
						//failed
						runb->state=4;

						RES=78;
						if(systemChecker.showLevel(RES, sysLog)){
							printf("D%-5d process ID:%d ",ID,runb->ID);
							systemChecker.check(RES, sysLog);
						}

						r=1;
					}//case r==-1 end
					break;
	
			case 0:	{
						//finished
						runb->state=3;
						runb->pptr->writeBack(*runb);
						finished.enQueue(runb->pptr);
						//finishedb.enQueue(runb);
						
						RES=177;
						if(systemChecker.showLevel(RES, sysLog)){
							printf("D%-5d process ID:%d ",ID,runb->ID);
							systemChecker.check(RES, sysLog);
						}

						pcbRelease(runb);
						//runb->release(); //return pcb to system
						//usedPcbNumber--;
						r=1;
					}//case r==0 end
					break;
	
			case 1:{
						//timp cycle arrives
						
						if(runb->interrupt==1){
							
							
							RES=189;
							if(systemChecker.showLevel(RES, sysLog)){
								printf("D%-5d process ID:%d ", ID,runb->ID);
								systemChecker.check(RES, sysLog);
							}

							readyb.insertToHead(runb); //insert runing process back to the head of ready queue, continue to run
							
													
							RES=188;
							if(systemChecker.showLevel(RES, sysLog)){
								printf("D%-5d process ID:%d ", ID,runb->ID);
								systemChecker.check(RES, sysLog);
							}
							break;
						}

						runb->state=0;
						//wait.enQueue(run);
						readyb.enQueue(runb);

						
						RES=171;
						if(systemChecker.showLevel(RES, sysLog)){
							printf("D%-5d process ID:%d ", ID,runb->ID);
							systemChecker.check(RES, sysLog);
						}
				   }//case r==1 end
					break;
	
			case 3:{	
						//system call return
						int mode; //whether wake up caller process
						runb->state=3;
						
						RES=174;
						if(systemChecker.showLevel(RES, sysLog)){
							printf("D%-5d system function ID:%d, caller process ID:%d ", ID,runb->ID,runb->callerID);
							systemChecker.check(RES, sysLog);
						}
						//pause[run->ID].deQueue(outPause);

						if((mode=IM.GM.read(10))==0 || mode==2){//0 means the sysfun will wake the process who calls it
																			//2 means the put the proc to the end of the ready queue
							
							//if(runb->ID==6) waitb[4].deQueue(outPauseb);	//P-operation is special, need to be changed
							waitb[runb->ID].deQueue(outPauseb);	//put sysfun caller process out of wait queue
				
							if(outPauseb->ID==runb->callerID){ //compare ID
								//	for(i=0; i<run->MData.mSize; i++)
								//		outPause->MData.write(i,run->MData.read(i));
								//wait.enQueue(outPause);			
							
								if(mode==0){
									readyb.insertToHead(outPauseb); //insert runing process back to the head of ready queue, continue to run
									RES=188;
									if(systemChecker.showLevel(RES, sysLog)){
										printf("D%-5d process ID:%d ", ID,outPauseb->ID);
										systemChecker.check(RES, sysLog);
									}
								}
								else{	//mode==2
										//in fact, i use this to simulate the condition that a process which has run a long time and  
										//has to be swaped to the ready queue, because my processes now are always very short, so i 
										//just simulate the real condition by this way ^_^
								
									readyb.enQueue(outPauseb);
									outPauseb->state=0;							
									printf("D%-5d process ID:%d, pause cause:%d ", ID,outPauseb->ID,runb->ID);
									RES=175;
									systemChecker.check(RES, sysLog);
								}

							}

							/*
							if(runb->pptr->MRgst.read(1)>=2){	//wake up a process
								pcbptr tmpPcbptr=0;
								waitb[4].get(tmpPcbptr,runb->pptr->MRgst.read(1)); //suspend in waitb[4] //need to be simplified!!!
								readyb.enQueue(tmpPcbptr);

								printf("D%-5d process ID:%d ",ID,tmpPcbptr->ID);
								RES=186;
								systemChecker.check(RES, sysLog);

								tmpPcbptr=0;
							}
							*/

						}else	//the sysfun wont wake up the caller processs, so the caller will be suspended
							if(mode==1){
								
								RES=185;
								if(systemChecker.showLevel(RES, sysLog)){
									printf("D%-5d process ID:%d, suspended process ID:%d ",ID, runb->ID, runb->pptr->callerID);
									systemChecker.check(RES, sysLog);
								}
							}	
				
						runb->pptr->MData.mem=0;	//destory the tem sysfun copy, set run->MData.mem=0 to avoid deconstruct caller's MData 
						delete runb->pptr;	//dont use delete runb

						pcbRelease(runb);
						//runb->release();	//return pcb to system
						//usedPcbNumber--;

						r=1;
				   }//case r==3 end
					break;

			case 5:{	
						//read system part 
						int address;	//system share operation start address
						int len;	//system share length
						int i;
							
						if(runb->priority!=0){	//only sysfun can visit this part
							//printf("error!");
						
							RES=180;
							if(systemChecker.showLevel(RES, sysLog)){
								printf("D%-5d process ID:%d, process priority:%d ", ID, runb->ID, runb->priority);
								systemChecker.check(RES, sysLog);
							}
						}
						else{
							address=IM.GM.read(11); //start address of system share to read or write
							len=IM.GM.read(12);	//the lenght to read or write

							//read system share to the lastest part of the MData 
							for(i=0; i<len; i++){ //eg. systemShare[0~2]={99,98,97},  then the last 3 unit of run->MData is 99,98,97
								runb->pptr->MData.write(runb->pptr->MData.mSize-len+i,systemShare[address+i]);
							}
					

							RES=181;
							if(systemChecker.showLevel(RES, sysLog)){
								printf("D%-5d process ID:%d ", ID,runb->ID);
								systemChecker.check(RES, sysLog);		
							}					
	
							readyb.insertToHead(runb); //insert runing process back to the head of ready queue, continue to run
							
														
							RES=188;
							if(systemChecker.showLevel(RES, sysLog)){
								printf("D%-5d process ID:%d ", ID,runb->ID);
								systemChecker.check(RES, sysLog);
							}
							//runb->state=0;	
							//readyb.enQueue(runb);	//mov the process into ready queue
							//printf("D%-5d process ID:%d ", ID,runb->ID);
							//RES=171;
							//systemChecker.check(RES, sysLog);

						}
						r=1;
					}//case r==5 end	
					break;

			case 6:{
						//write system part 
						int address;	//system share operation start address
						int len;	//system share length
						int i;
							
						if(runb->priority!=0){	//only sysfun can visit this part
							//printf("error!");
							
							RES=180;
							if(systemChecker.showLevel(RES, sysLog)){
								printf("D%-5d process ID:%d, process priority:%d ", ID, runb->ID, runb->priority);
								systemChecker.check(RES, sysLog);
							}
						}
						else{
							address=IM.GM.read(11); //start address of system share to read or write
							len=IM.GM.read(12);	//the lenght to read or write

							//write lastest part of the MData into system share
							for(i=0; i<len; i++){
								systemShare[address+i]=runb->pptr->MData.read(runb->pptr->MData.mSize-len+i);
							}
							
							RES=182;
							if(systemChecker.showLevel(RES, sysLog)){
								printf("D%-5d process ID:%d ", ID,runb->ID);
								systemChecker.check(RES, sysLog);		
							}
								
							readyb.insertToHead(runb); //insert runing process back to the head of ready queue, continue to run
													
													
							RES=188;
							if(systemChecker.showLevel(RES, sysLog)){
								printf("D%-5d process ID:%d ", ID,runb->ID);	
								systemChecker.check(RES, sysLog);
							}

							//runb->state=0;
							//readyb.enQueue(runb);	//mov the process into ready queue
							//printf("D%-5d process ID:%d ", ID,runb->ID);
							//RES=171;
							//systemChecker.check(RES, sysLog);
						}
						r=1;	
					}//case r==6 end		
					break;

			case 7:{
						//wake up process
						int waked;	//communicate of the process who need to be waked 
						int where;	//which queue to wake
						pcbptr tmpPcbptr=0;	//temp pcb pointer

						if(runb->priority!=0){	//only sysfun can visit this part
							//printf("error!");
							
							RES=187;
							if(systemChecker.showLevel(RES, sysLog)){
								printf("D%-5d process ID:%d, process priority:%d ", ID, runb->ID, runb->priority);
								systemChecker.check(RES, sysLog);
							}
						}
						else{
							waked=IM.GM.read(11);	//read communicate
							where=IM.GM.read(12);	//read wait queue number

							if(where>=20 || where<0){	//over the range
								
								RES=190;
								if(systemChecker.showLevel(RES, sysLog)){
									printf("D%-5d process ID:%d requested wait queue NO:%d ",ID,runb->ID,where);
									systemChecker.check(RES, sysLog);
								}
							}
							else
								if(waitb[where].CommunicateGet(tmpPcbptr,waked)!=-1){ //suspend in waitb[4] //need to be simplified!!!
									readyb.enQueue(tmpPcbptr);	//wake up the process and let it be ready

									
									RES=186;
									if(systemChecker.showLevel(RES, sysLog)){
										printf("D%-5d process ID:%d, waked process ID:%d ",ID,runb->ID,tmpPcbptr->ID);
										systemChecker.check(RES, sysLog);
									}
	
									tmpPcbptr=0;
								}

							readyb.insertToHead(runb); //insert runing process back to the head of ready queue, continue to run
							
														
							RES=188;
							if(systemChecker.showLevel(RES, sysLog)){
								printf("D%-5d process ID:%d ", ID,runb->ID);
								systemChecker.check(RES, sysLog);
							}
							//runb->state=0;
							//readyb.enQueue(runb);	//mov the process into ready queue
							//printf("D%-5d process ID:%d ", ID,runb->ID);
							//RES=171;
							//systemChecker.check(RES, sysLog);
		
						}
						r=1;
					}//case r==7 end
					break;

			default:{
						//system call
						
						RES=173;
						if(systemChecker.showLevel(RES, sysLog)){
							printf("D%-5d caller process ID:%d, system call NO:%d ",ID,runb->ID,r-10); //syscall checked, at this time r-10 == sysfun ID
							systemChecker.check(RES, sysLog);
						}

						sysCall.copy(r-10,sysTem); //got system fun according to r
						//for(i=0; i<run->MData.mSize; i++){
						//	sysTem->MData.write(i,run->MData.read(i)); //copy MData to syscall ??system will change waitqueue
						//}
						//not copy MData!!! use the way as follows:
						sysTem->MData.mem=runb->pptr->MData.mem;	//sysTem->MData.mem pointer to the same unit as caller
						sysTem->MData.mSize=runb->pptr->MData.mSize;	//make sysTem->MData same mSize as caller

						sysTem->callerID=runb->ID; //indicate who calls sysTem
						//pause[r-10].enQueue(run);	//proc to wait sysfun completed //simplified!!! //have changed
						
						//if(r==16) waitb[4].enQueue(runb);	//P-operation is speciall
						waitb[r-10].enQueue(runb);	//proc to wait sysfun completed //simplified!!! //have changed

						runb->state=2;
						
						RES=172;
						if(systemChecker.showLevel(RES, sysLog)){
							printf("D%-5d process ID:%d, pause cause:%d ",ID,runb->ID,sysTem->ID);	//casue need be to changed
							systemChecker.check(RES, sysLog);
						}

						loaderSysfun(sysTem);	//loader system function to a pcb and add to ready queue
							
									
						RES=188;
						if(systemChecker.showLevel(RES, sysLog)){
							printf("D%-5d process ID:%d ", ID,sysTem->ID);
							systemChecker.check(RES, sysLog);
						}
			
						//sysTem->state=0;
						//printf("D%-5d system function ID:%d, caller ID:%d ", ID,sysTem->ID,sysTem->callerID);
						//RES=176;
						//systemChecker.check(RES, sysLog);
					//	sysTem->~process();
						r=1;
					}//default end
		
		}//switch end
	
	}
	
	RES=179; //queue is empty is normal when scheduler finished
	if(systemChecker.showLevel(RES, sysLog)){
		printf("D%-5d interpreter ID:%d ",ID, IM.ID);
		systemChecker.check(RES, sysLog);
	}
	return 0;
}

int dispatcher::pcbManagement(){	//manage pcb
	int i=0;
	while(i<pcbNum && pcbSet[i].valid!=0) i++;
	if(i==pcbNum){
		
		RES=183; //all pcb are used now
		if(systemChecker.showLevel(RES, sysLog)){
			printf("D%-5d ",ID);
			systemChecker.check(RES, sysLog);
		}

		return -1;
	}
	else{
		pcbSet[i].valid=1; //this pcb is used now
		usedPcbNumber++;
		return i;
	}
}

void dispatcher::pcbRelease(pcbptr pblock){
	pblock->ID=pblock->state=pblock->priority=pblock->callerID=pblock->communicate=0;
	pblock->valid=0;
	pblock->pptr=0;
	pblock->next=0;
	usedPcbNumber--;
}

//int dispatcher::setProperty(int term, int value){
//	switch(term){
//		case 0:  
//	}
//}

