#include "component.h"

int config(void){	//读取系统的配置信息
	FILE *fp;
	char para[40];

	if((fp=fopen("config.ini","r"))==NULL)
		return -1;
	do{
		fscanf(fp,"%s",para);
		if(!strcmp(para,"end")) break;
		
		char *cp=para;
		int r;
		while(*cp!='=') cp++;
		
		*cp=0;
		r=atoi(cp+1);
		if(!strcmp(para,"delayMode"))	//延迟方式
			delayMode=r;
		if(!strcmp(para,"displayMode"))	//显示模式
			displayMode=r;
		if(!strcmp(para,"reportLevel"))	//运行情况报告
			reportLevel=r;
		if(!strcmp(para,"updateSysfun"))	//重新汇编系统程序库
			updateSysfun=r;

	}while(1);

	return 0;
}


int main(){

	if(config()) printf("system parameters config error!\n");

	/*
	delayMode=0;
	displayMode=1;	//determine information dispay mode
					//0 hide the content of the process loaded to system; 1 show all information 
	reportLevel=2; //0 report Warning and Error, 1 report Normal, Warning and Error, 2 report all
	updateSysfun=0;
	*/
	
	interpreter IM(101,30);
	editor e(103);
	assembler a(102);
	dispatcher os(105,10,30,40);
	storage  sys(106, sysPath[0]);
	storage  disk(107,sysPath[1]);


	//************************************************************************************************************
	//Dont modify this part!
	{	//initialize system
		FILE *fp;	//file pointer used to open code file for system function 
		int i;
		procptr pptr;
		pptr=new process[sysFunNumber];	//create process for system function
		process tp(0,100,10,40,0);	//process template, the code part is large but the data part is very small, with high priority
		
		for(i=0; i<sysFunNumber; i++){
			pptr[i].copy(tp);	//copy template process to each sysfun process
			pptr[i].ID=i;	//give each sysfun a new id(system function call number) 
			
		
			if(updateSysfun==1){	//reassemble sysfun
				sys.getFile(fp, sysFunTable[i],0); //open the function code file
				if(e.ASM(fp,a,sysFunTable[i])!=-1){
					Load(pptr,sysFunTable[i]);
					os.sysCall.put(i,&pptr[i]);	//load the system function to the sysCall list in scheduler
				}
				sys.releaseFile(fp);	//close file
			}
			else{
				Load(pptr,sysFunTable[i]);
				os.sysCall.put(i,&pptr[i]);	//load the system function to the sysCall list in scheduler
			}
	

		}
	}	//throuth this step, all the system function has been load into the system
	//************************************************************************************************************

	//************************************************************************************************************
	{	//load user code to system
		FILE *run;	//file ptr of list file
		FILE *pfr;	//pointer to the code file
		procptr pptr;	//used to create user process
		process tpf(0,100,40,50,1);	//process template, the data part is larger than sysfun, but priority is lower

		int id=50;	//least user proc ID
		char name[50]; //used to code file name
		disk.getFile(run,runList,0);
		fscanf(run,"%s",name);	//read code file name
		while(strcmp(name,"end") && strcmp(name, "END")){ //list file end with "end" or "END"
			disk.getFile(pfr,name,0);	//open code file
			pptr=new process;	//create user process
			pptr->copy(tpf);	//copy template process to each process
			pptr->ID=id++;	//give user process a new id
			printf("process ID:%d\n", pptr->ID);
	
			if(e.ASM(pfr,a,name) != -1){
				Load(pptr,name);
				os.loader(pptr);	//load process to scheduler
			}
			disk.releaseFile(pfr);	//close code file 
			fscanf(run,"%s",name); //read next code file name
		}
		disk.releaseFile(run);	//close list file
	}	//throuth this step, all the user process has been load into the system
	//*************************************************************************************************************

	
	os.swap2(IM);	//swap process to run and wait, when a user process finished, it will be put into finished queue 
	

	//*************************************************************************************************************
	{	//show the finished process data part
		procptr pptr;
		//pcbptr pblock;
		int r=0;
		//r=os.finished.deQueue(pptr); //get a process from finished queue
		r=os.finished.deQueue(pptr);
		
		
		printf("\ndisplay data?(y/n): ");
		char c;
		c=getche();
		putchar('\n');

		
		while(r!=-1){	//finished queue isnt empty
			printf("process ID:%d\n", pptr->ID);
			if(c=='y' || c=='Y'){
				int i;
				printf("memory ID:%d\n",pptr->MData.ID);
				for(i=0; i<pptr->MData.mSize; i++)
				printf("%-4d%d\n",i,pptr->MData.read(i)); //show M.mem[i]
			}
			printf("execute times:%d\n",pptr->exetime);
			r=os.finished.deQueue(pptr);	//get next process
		}
	}
	//*************************************************************************************************************

	//systemChecker.check(RES, sysLog);	//show system end state, normal state is show "queue::deQueue: queue is empty"

	printf("\nPress any key to exit.");
	getch();

	return 0;
}

