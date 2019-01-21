#include <stdio.h>
#include <iostream.h>
#include <string.h>
#include <stdlib.h>
#include <conio.h>

struct abendNote;	//系统运行状态结构
struct varNote;		//符号表项结构，汇编时使用
	struct varNote2;	//标号表项/跳转语句表项结构，是符号表项子类，汇编时使用
class assembler;	//汇编工具类
class checker;		//检错类，用于检测系统状态
class editor;		//编辑器类，使用汇编工具现行汇编
class interpreter;	//解释器类，用于解释执行指令
class list;			//链表类
	class sysList;	//系统函数链表类，是链表类子类用来存储指令和数据
class memory;		//存储器类，
	class stack;	//堆栈类，存储器类子类
class process;		//进程类
class queue;		//队列类
class pcb;			//进程控制块类
class pcbList;		//进程控制块链表类，动态链表

typedef process* procptr; //process pointer
typedef pcb* pcbptr;	//pcb pointer

//以下全局变量在global.cpp中定义
extern int RES;
extern checker systemChecker; 
extern abendNote sysLog[100];
extern int sysFunNumber;
extern char* sysFunTable[10];
extern char* sysPath[2];
extern char runList[40];
extern int displayMode;
extern int reportLevel;
extern int delayMode; //0 Load/Store无延时；1 Load/Store有延时
extern int updateSysfun;
extern void procDelay(int mode, long time_ms);
extern int Load(procptr pptr, char *s);	//临时读字符码到内存函数

struct abendNote{
	int abendCode;	//abend NO.
	int level;	//Normal 0, Sub conditon 1, Warning 2, Error 3, 
	char abendReason[150];	//abend reason
};

struct varNote{	//note the var delcared in code
	int valid;	//the var has exited or not
	char var[10];	//var name with max length 10
	int  pos;	//var address in memory
	varNote() { valid=pos=0; };
};

struct varNote2: public varNote{
	int found;
	varNote2():varNote(){
		found=0;
	};
};

class checker{	//report the running condition
public:
	int ID;
	checker(int key){ ID=key; };
	void check(int k,abendNote *abendTable);
	int showLevel(int k, abendNote *abendTable);	//whether show the report or not

};

class memory{ //友元类 interpreter，process，dispatcher
//protected: //for process to copy; need to improve!!!
protected:
	int *mem;	//ptr to storage units 
public:
	//int *mem;	//ptr to storage units 
	int ID;     //handle of the memory
	int mSize;	//size of the memory
	//int *ptrTable;	//pointer var Table

	memory() { ID=0; mem=0; }; //overload
	memory(int key, int m=20):mSize(m){ 
		int i; 
		ID=key;
		mem=new int[m];
		//ptrTable=new int[m];
		for(i=0; i<m; i++){	//initialize the memory with size of mSize and value of zero
			*(mem+i)=0;
			//*(ptrTable+i)=0;
		}
	};	//constructor

	void update(int m);
	~memory(){
		if(mem) {delete mem; mem=0;}
		//if(ptrTable)	{delete ptrTable; ptrTable=0;}
	};	//deconstructor

	int read(int i);
	int write(int i, int k);
	//int editor(void);	//use to edit MData
	//nt editor2(void);	//use to edit MCode
	//int editorFromFile(FILE *pfile); //use to edit input file
	//int editorToFile(FILE *pfile); //use to edit to file

	friend class interpreter;
	friend class dispatcher;
	friend class process;
};

class stack: public memory{
	int SP;	//stack pointer
public:
	
	stack() { SP=0; };
	stack(int k, int m=10):memory(k,m){
		SP=0;
	};
//	~stack() { ; };

	int push(int i);
	int pop();
};

class assembler{
public:
	int ID;
	assembler(int key){ ID=key; };
	//int trans(char *op, varNote *vp, int &ptrflag,int line);
	int trans2(char *op, varNote *vp, int line,int &addressing);
	int trans(char *op, int line);
};

class process{
public:
	int ID;
	int state;	//0 ready, 1 run, 2 wait, 3 finished, 4 failed
	int exetime; //0 new
	int priority; //default 0;
	int callerID; //ID of who calls the proc
	int communicate; //used when commucate with anohter prcess;
	int interrupt;	//0, allow process be interupted; 1, dont allow interruption
	
	int update;	//flag about whether some property of a process is renewed. eg. communicate changed

	memory MCode;
	memory MData;
	//memory MRgst;
	stack  S;
	
	process(){ 
		ID=0; 
		priority=0; 
		state=0; 
		exetime=0; 
		callerID=0; 
		communicate=0; 
		update=0; 
		interrupt=0;
	};
	
	~process() { MCode.~memory(); MData.~memory(); S.~stack(); };

	process(int pid, int csz=40, int dsz=20, int ssz=10, int prty=0):
	MCode(10*pid,csz),MData(10*pid+1,dsz),S(10*pid+3,ssz)	{ //ID of memory and stack is assigned by process ID
		ID=pid;
		priority=prty;
		state=0;
		exetime=0;
		callerID=0;
		communicate=0;
		update=0;
		interrupt=0;
	};
	
	process(int pid, int way, process &sproc){ //copy conductor //way!=1 ordinary ;way=1 sysfun conductor
		int i;

		*this=sproc;
		ID=pid;	
	
		MCode.mem= new int[sproc.MCode.mSize];
//		MCode.ptrTable= new int[sproc.MCode.mSize];

		if(way!=1){
			MData.mem= new int[sproc.MData.mSize];
//			MData.ptrTable= new int[sproc.MData.mSize];
		}
		else{
			MData.mem=0;
//			MData.ptrTable=0;
			MData.mSize=0;
		}

		//MRgst.mem= new int[sproc.MRgst.mSize];
		//MRgst.ptrTable= new int[sproc.MRgst.mSize];

		S.mem= new int[sproc.S.mSize];
//		S.ptrTable= new int[sproc.S.mSize];

		for(i=0; i<sproc.MCode.mSize; i++){	//copy MCode
			MCode.mem[i]=sproc.MCode.mem[i];
//			MCode.ptrTable[i]=sproc.MCode.ptrTable[i];
		}
		
		if(way!=1)
			for(i=0; i<sproc.MData.mSize; i++){	//copy MData
				MData.mem[i]=sproc.MData.mem[i];
//				MData.ptrTable[i]=sproc.MData.ptrTable[i];
			}

		//for(i=0; i<sproc.MRgst.mSize; i++){	//copy MRgst
		//	MRgst.mem[i]=sproc.MRgst.mem[i];
		//	MRgst.ptrTable[i]=sproc.MRgst.ptrTable[i];
		//}

		for(i=0; i<sproc.S.mSize; i++){		//copy S
			S.mem[i]=sproc.S.mem[i];
//			S.ptrTable[i]=sproc.S.ptrTable[i];
		}
	};

	void copy(process &sproc); //copy 
	void writeBack(pcb &block); //write the pcb conetent back to process

};

class editor{
	const int codeBufferSize; //代码缓冲区的大小
	int codeBuffer[200];	//用于临时保存代码段
	int dataBuffer[100];	//用于临时保存数据段
public:
	int ID;

	editor(int key, int codeSize=200): codeBufferSize(codeSize){ 
	//	int i;
		ID=key; 
	//	for(i=0; i<codeBufferSize; i++){	//初始化缓冲区
	//		codeBuffer[i]=0;
	//		dataBuffer[i]=0;
	//	} //未知问题
	};

	//int editorCollection(memory &M, int k,assembler &A);	//choose different editor	
	//int editorCollection(memory &M, FILE *pfile, int k, assembler &A);	//choose different editor	
	//int editorData(memory &M, assembler &A);	//use to edit MData
	int editorDataFromFile(FILE *pfile, memory &M, assembler &A); ////use to edit input data file
	//int editorCode(memory &M, assembler &A);	//use to edit MCode
	//int editorCodeFromFile(FILE *pfile, memory &M, assembler &A); //use to edit input code file
	int editorFromFile(FILE *pfile, assembler &A, int &dataNumInMData); //use to edit input code file //返回代码段长度
	//int editorFromFileToFile(FILE *pfile, procptr pptr, assembler &A); //use to edit input code file
	//int editorCodeToFile(FILE *pfile, memory &M, assembler &A); //use to edit to file

	int ASM(FILE *pfile, assembler &A, char *s); //汇编程序
	/*
	void saveMem(FILE *fp, int len){	//将代码段和数据段的内容保存到文件
		int i;
		for(i=0; i<len; i++){
			fprintf(fp,"%d ",ad(i));
		}
	};
	*/
	
};

/*
class processor{
public:
	int ID;
	int flag;	//state register
	int CM;	//counter in code memory
	processor(int key) { ID=key; flag=0; CM=0; };
	
	void offset(memory &M, int id, int is) {	M.write(id, M.read(id+M.read(is)));	};	//mov (id+(is)) to id
	void mov(memory &M, int id, int is) { M.write(id, M.read(is)); };	//mov index is to index id in memory M
	void add(memory &M, int id, int is) { M.write(id, M.read(id) + M.read(is)); };	//add index is to index id in memory M
	void sub(memory &M, int id, int is){	//sub index is to index id in memory M
		int r=0;
		r=M.read(id) - M.read(is);
		M.write(id, r);
		if(r>0) flag=1;	//set the state register
		else
			if(r<0) flag=-1;
			else flag=0;
	};	
	void inc(memory &M, int id) { M.write(id, M.read(id) + 1); }; //add 1 to id
    void cmp(memory &M, int id, int is){	//cmp the value between is and id
		int r=0;
		if((r=M.read(id)-M.read(is))>0)	flag=1;
		else 
			if(r<0) flag=-1;
			else flag=0;
	};	
	void ja(int line) { CM=4*line; flag=0; }; //jmp when flag>0; the real line NO. in memory is 3 times than on screen
	void jb(int line) { CM=4*line; flag=0; }; //jmp when flag<0
	void je(int line) { CM=4*line; flag=0; }; //jmp when flag==0
	void jne(int line) { CM=4*line; flag=0; }; //jmp when flag!=0
	void jmp(int line) { CM=4*line; flag=0; }; //jmp whenever

    void push(memory &M, int id, stack &S) { S.push(M.read(id)); }; //push to stack S
	void push(int vs, stack &S) {S.push(vs); }; //overload immediate push
	void pop(memory &M, int id, stack &S) { M.write(id, S.pop()); }; //pop from stack S
	void pop(int &vs, stack &S) { vs=S.pop(); }; //immediate pop
	void mov(memory &M1, int id, memory &M2, int is) { M1.write(id, M2.read(is)); };	//mov index is in M2 to index id in M1
    void add(memory &M1, int id, memory &M2, int is) { M1.write(id, M1.read(id) + M2.read(is)); };	//add index is in M2 to index id in M1
	void cmp(memory &M1, int id, memory &M2, int is){	//overload cmp in 2 memories
		int r=0;
		if((r=M1.read(id)-M2.read(is))>=0)	flag=1;
		else
			if(r<0) flag=-1;
		else flag=0;
	};
	
	void Immov(memory &M, int id, int vs) { M.write(id, vs); };	//immediate mov
	void Imadd(memory &M, int id, int vs) { M.write(id, M.read(id) + vs); };	//immediate add	
	void Imsub(memory &M, int id, int vs){	//immediate sub
		int r=0;
		r=M.read(id) - vs;
		M.write(id, r);
		if(r>0) flag=1;
		else
			if(r<0) flag=-1;
			else flag=0;
	};	
	void Imcmp(memory &M, int id, int vs){	//Immediate cmp
		int r=0;
		if((r=M.read(id)-vs)>0)	flag=1;
		else 
			if(r<0) flag=-1;
			else flag=0;
	};
	
};
*/


class interpreter{
	int cycleTimes;	//max cycletimes at a execute time
	memory MRgst; //内部寄存器，用于寄存操作码，操作数的地址等	
	memory GM;	//general register group，通用寄存器组
	int GMNumber;	//通用寄存器的数目，即通用寄存器组的大小
	int flag;	//状态字寄存器，目前只保存cmp指令的比较结果（正1，负-1，零0）
	int PC;	//指令计数器
public:
	int ID;

	interpreter(int key, int cc=20, int regNum=10, int GMNum=20):MRgst(10*key, regNum),GM(10*key+1,GMNum){
		ID=key;
		cycleTimes=cc;
		GMNumber=GMNum;
	};
	int exer(process &proc);
	int effectAddressing(int addcode, int fadd, process &proc, int &RorM);//根据寻址方式和形式地址，计算有效地址
	void load(memory &M1, int id, memory &M2, int is) { procDelay(delayMode,100); M1.write(id, M2.read(is)); };	//mov index is in M2 to index id in M1
	void store(memory &M1, int id, memory &M2, int is) { procDelay(delayMode,100); M1.write(id, M2.read(is)); };	//mov index is in M2 to index id in M1
	void push(int vs, stack &S) { procDelay(delayMode,50); S.push(vs); }; //immediate push
	void pop(int &vs, stack &S) { procDelay(delayMode,50); vs=S.pop(); }; //immediate pop

	friend class dispatcher;
};

/*
class userInterface{	//UI
public:
	int ID;
	FILE *f1;
	userInterface(int key) { ID=key; };
	//int menu(processor &P1, interpreter &I1, process &proc, checker &CK, editor &e, assembler &a);	
	int showMemory(memory &M);
};
*/

class queue{	//unit in queue is pointer to process
	procptr *procQueue;
	int head;	//head pos
	int tail;	//tail pos
	int memberNumber; //numbers of members in queue, max memberNumber=length-1!!!
public:
	int ID;
	int length;	//length

	queue(int key=0, int len=10){ //notice the id!!! can has default
		memberNumber=0;
		ID=key;
		head=tail=0;
		length=len;
		procQueue=new procptr[len];
	};
	~queue(){
		if(procQueue!=0) { delete procQueue; procQueue=0; }
	};
	
	int enQueue(procptr pptr);
	int deQueue(procptr &pptr);

};

class list{
protected:
	procptr *listp;
public:
	int ID;
	int length;	//length
	
	list(int key, int len=10){
		int i;
		ID=key;
		length=len;
		listp=new procptr[len];
		for(i=0; i<len; i++){
			listp[i]=0;
		}//initialize lisp
	};

	~list(){
		if(listp!=0) { delete listp; listp=0; }
	};

	int put(int t, procptr pptr);
	int get(int t, procptr &pptr);

};

class sysList: public list{
public:
	sysList(int key, int len=5):list(key,len){ ; };
	int copy(int t, procptr &dpptr);
	int copy(procptr &dpptr, procptr spptr);
};

/*
class scheduler{
public:
	int ID;
	procptr sysTem;//only one sysfun can run

	procptr run;	//ptr to the process is running
	queue wait;	//wait queue
	queue pause[10];	//stop queue, it has 10 condition to cause process to pause
	queue finished;

	int systemShare[200];

	sysList sysCall;	//sysfun list
	//list sysCall;
	int priorityCounterTable[3]; //times of execute interpreter exer times

	scheduler(int key, int waitLen=10, int sysFunLen=10, int finishLen=20):wait(key*10, waitLen),sysCall(key*10+2,sysFunLen),
	finished(key*10+3,finishLen)
	{
		ID=key;
		run=0;
		sysTem=0;
		priorityCounterTable[0]=10;
		priorityCounterTable[1]=5;
		priorityCounterTable[2]=2;
	};
	void loader(procptr pptr); //load proc(pointer) to wait queue
	int swap(interpreter &IM, processor &PM);	//swap process to run and wait
	int swap2(interpreter &IM, processor &PM);	//not copy data memory, but use pointer to realize
};
*/
class storage{
public:
	char pos[50];
	char visitWay[5]; //0 read, 1 write, 2 read update
	int ID;
	storage(int key) { ID=key; };
	storage(int key, char *store){
		ID=key;
		strcpy(pos,store);
	};
	int getFile(FILE* &pf, char *filename, int choice);	
	int releaseFile(FILE* &fp);	//close file
};

class pcb{
	int state;	//process state
	int priority;	//process priority
	int exetime;	//process exetime
	int callerID;	//process caller ID
	int communicate;
	int	interrupt;
	int	valid;	//0, pcb is free; 1, pcb is used
public:
	//static int total;	//total pcb num
	
	int ID;	//pcb ID is same with process ID ??


	procptr pptr;	//point to a process
	pcb* next;	//point to next process control block

	pcb(){	//blank pcb;
		ID=callerID=-1;
		state=priority=exetime=0;
		pptr=0;
		next=0;
		communicate=0;
		interrupt=0;
		valid=0;
	};

	pcb(process &proc){	//initialize a pcb with another pcb
		//total++;
		ID=proc.ID;
		state=proc.ID;
		priority=proc.priority;
		exetime=proc.exetime;
		callerID=proc.callerID;
		communicate=proc.communicate;
		interrupt=proc.interrupt;
		pptr=&proc;
		next=0;
	};

	void update(procptr pptr); //update a pcb when some properties of its process is changed
	void copy(pcb &block);	//copy pcb to current pcb
	void link(process &proc);	//link a process to a pcb and use the properties of the process to change the pcb
	void release(); //return the pcb to system and let other process can use it

	friend class dispatcher;
	friend class process;
	friend class pcbList;
};

class pcbList{
	pcbptr head;	//head pointer of a list
	pcbptr tail;	//tail pointer of a list
	pcb blank;	//the first pcb is a blank pcb as a filler, used to make it convenient for list operation,sometime op is different to the first node
public:
	int ID;		
	int len;	//length of the list
	
	pcbList(int id=20){
		ID=id;
		head=&blank;
		tail=&blank;
		len=0;	//blank not included in pcb
	};

	int enQueue(pcbptr pblock);	//insert the block to the tail of the list by refer its ptr
	int deQueue(pcbptr &pblock);	//return the ptr of the first block which is deleted from the list
	int IDGet(pcbptr &pblock,int id);	//get pcb according to its id	
	int CommunicateGet(pcbptr &pblock,int com);	//get pcb according to its communicate
	int insertToHead(pcbptr pblock);	//insert a pcb to the head of a list
};

//new scheduler
class dispatcher{
	
	const int pcbNum;	//max pcb number
	pcb* pcbSet;
	int usedPcbNumber;
	procptr sysTem;//only one sysfun can run
	pcbptr runb;
	pcbList readyb;	//ready queue
	pcbList waitb[20]; //wait queue(notice: in old scheduler wait queue is redy queue, pause queue is wait queue

	int systemShare[200];	//system share region
	int priorityCounterTable[3]; //times of execute interpreter exer times
	

public:
	int ID;
	//pcbList finishedb;	//finished queue
	queue finished;
	sysList sysCall;	//sysfun list

	//procptr run;	//ptr to the process is running
	//queue wait;	//wait queue	
	//queue pause[10];	//stop queue, it has 10 condition to cause process to pause
	//list sysCall;

	dispatcher(int key, int sysFunLen=10, int finishLen=30, int pcbs=20):pcbNum(pcbs),sysCall(key*10+2,sysFunLen),finished(key*10+3,finishLen)
	{
		ID=key;
		runb=0;
		sysTem=0;
		priorityCounterTable[0]=10;
		priorityCounterTable[1]=5;
		priorityCounterTable[2]=2;

		pcbSet=new pcb[pcbNum];	
		usedPcbNumber=0;
	};
	
	//void loader(procptr pptr); //load proc(pointer) to wait queue
	void loader(procptr pptr);	//new loader, load a process to a pcb and add the pcb to ready queue
	void loaderSysfun(procptr pptr);	//sysfun loader, load a sysfun to a pcb and add the pcb to head of ready queue
	int swap2(interpreter &IM);	//not copy data memory, but use pointer to realize
	int pcbManagement(); //managepcb
	void pcbRelease(pcbptr pblock); //release a pcb
	//int setProperty(int term, int value); //set process property;

	void tool_pcbStatistics();	//check the num of pcb;
};