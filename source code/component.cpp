#include "component.h"

void memory::update(int m){ 
	int i; 		
	delete mem;
	//delete ptrTable;
	mSize=m;
	mem=new int[m];
//	ptrTable=new int[m];
	for(i=0; i<m; i++){	//initialize the memory with size of mSize and value of zero
		*(mem+i)=0;
//		*(ptrTable+i)=0;
	}
	
	RES=18;
	if(systemChecker.showLevel(RES, sysLog)){
		printf("D%-5d new size:%d ",ID,m);
		systemChecker.check(RES, sysLog);
	}
}

int memory::read(int i){ 
	if(i>=mSize || i<0){
		
		RES=10;
		if(systemChecker.showLevel(RES, sysLog)){
			printf("D%-5d memory size:%d ,read address:%d ",ID,mSize,i);	//overflow
			systemChecker.check(RES, sysLog);
		}
		return -1;
	}
	else
		return mem[i]; 
}	//read from memory

int memory::write(int i, int k){ 
	if(i>=mSize || i<0){
		
		RES=11;
		if(systemChecker.showLevel(RES, sysLog)){
			printf("D%-5d memory size:%d ,write address:%d ",ID,mSize,i);   //overflow
			systemChecker.check(RES, sysLog);
		}
		return -1;
	}
	else{
		mem[i]=k; 
		return 0;
	}
}	//write to memory

int stack::push(int i){ 
	if(SP<mSize){	//check whether the stack is full
		write(SP,i);
		SP++; //SP points to the next unit of current data
		return 0;
	}
	else{
	
		RES=16;
		if(systemChecker.showLevel(RES, sysLog)){	
			printf("D%-5d stack size:%d, SP:%d ",ID,mSize,SP);
			systemChecker.check(RES, sysLog);
		}
		return -1;
	}
}

int stack::pop(){
	int r; 
	if(SP>0){	//check whether the stack is empty
		SP--; 
		r=read(SP); 
		return r;
	}
	else{
		
		RES=17;
		if(systemChecker.showLevel(RES, sysLog)){	
			printf("D%-5d stack size:%d, SP:%d ",ID,mSize,SP);
			systemChecker.check(RES, sysLog);
		}
		return -1;
	}		
}

/*
int assembler::trans(char *op, varNote *vp, int &ptrflag,int line){	//ptrflag = 1 means op is ptr var
	int i=0;
	char tmp[10];

	if(op[0]=='@'){	//the var is a ptr
		strcpy(tmp, op+1);	//judge from 2nd character
		strcpy(op,tmp);
		ptrflag=1;	//indicate that this var is ptr
	}

	while(vp[i].valid==1){
		if(!strcmp(op,vp[i].var))	//the var has been noted before
			return vp[i].pos;
		i++;
	}
	
	if(ptrflag==1){	//systax error use '@' but not point to any existed var
		
		RES=51;
		if(systemChecker.showLevel(RES, sysLog)){
			printf("D%-5d row:%d ",ID,line);
			systemChecker.check(RES, sysLog);
		}
		return -1;
	}
	else{
		if(!strcmp(op,"$")) return 0;	//filler
		if(op[0]=='-' || (op[0]>='0' && op[0]<='9'))
			return atoi(op);	//immediate num
		
		RES=52;
		if(systemChecker.showLevel(RES, sysLog)){
			printf("D%-5d row:%d, indentifier:\'%s\' ",ID,line,op);
			systemChecker.check(RES, sysLog);
		}
		return -1; //error identifier
	}
}
*/

void process::copy(process &sproc){ //copy 
	int i;
	*this=sproc;
	
	MCode.mem= new int[sproc.MCode.mSize];
//	MCode.ptrTable= new int[sproc.MCode.mSize];

	MData.mem= new int[sproc.MData.mSize];
//	MData.ptrTable= new int[sproc.MData.mSize];

	//MRgst.mem= new int[sproc.MRgst.mSize];
	//MRgst.ptrTable= new int[sproc.MRgst.mSize];

	S.mem= new int[sproc.S.mSize];
//	S.ptrTable= new int[sproc.S.mSize];

	for(i=0; i<sproc.MCode.mSize; i++){	//copy MCode
		MCode.mem[i]=sproc.MCode.mem[i];
//		MCode.ptrTable[i]=sproc.MCode.ptrTable[i];
	}
		
	for(i=0; i<sproc.MData.mSize; i++){	//copy MData
		MData.mem[i]=sproc.MData.mem[i];
//		MData.ptrTable[i]=sproc.MData.ptrTable[i];
	}

	//for(i=0; i<sproc.MRgst.mSize; i++){	//copy MRgst
	//	MRgst.mem[i]=sproc.MRgst.mem[i];
	//	MRgst.ptrTable[i]=sproc.MRgst.ptrTable[i];
	//}

	for(i=0; i<sproc.S.mSize; i++){		//copy S
		S.mem[i]=sproc.S.mem[i];
//		S.ptrTable[i]=sproc.S.ptrTable[i];
	}
}

void process::writeBack(pcb &block){	//write the pcb content back to process
	state=block.state;
	exetime=block.exetime;
}

/*
int editor::editorCollection(memory &M, int k,assembler &A){	//choose different editor
	if(k==0){
		editorData(M,A);	//MData editor
		return 0;
	}
	if(k==1){
		editorCode(M,A);	//MCode editor	
		return 0;
	}
	
	RES=19;
	if(systemChecker.showLevel(RES, sysLog)){
		printf("D%-5d call num:%d ",ID,k);
		systemChecker.check(RES, sysLog);
	}
	return -1;
};
*/
/*	
int editor::editorCollection(memory &M, FILE *pfile, int k, assembler &A){	//choose different editor
	if(k==0){
		editorCodeFromFile(pfile,M,A);	//edit code from file
		return 0;
	}
	if(k==1){
		editorCodeToFile(pfile,M,A);	//edit code and save to file
		return 0;
	}
	
	RES=19;
	if(systemChecker.showLevel(RES, sysLog)){
		printf("D%-5d call num:%d ",ID,k);
		systemChecker.check(RES, sysLog);
	}
	return -1;
}
*/
/*
int editor::editorData(memory &M, assembler &A){	//edit the content of memory at one time
	char r[10];
	int i=0;
	printf("enter the content in this memory ID:%d\n",M.ID);
	printf("%-4d",i);	//num of line
	scanf("%s",r);
	getchar();
	while((r[0]=='-' || (r[0]>='0' && r[0]<='9')) && i<M.mSize){	//end edit
		M.write(i,atoi(r));	//trans r to num
		i++;
		printf("%-4d",i);
		scanf("%s",&r);
		getchar();
	}
	if(!strcmp(r,"END") || !strcmp(r,"end")){
		printf("end edit\n");
		return 0;
	}
	else{
		
		RES=12;
		if(systemChecker.showLevel(RES, sysLog)){
			printf("D%-5d row:%d, data:\"%s\" ",ID,i,r);
			systemChecker.check(RES, sysLog);
		}
		return -1;
	}
}
*/

/*
int editor::editorCode(memory &M, assembler &A){	//editor in line
	int o,d,s;	//opcode, destination operand and source oprand
	char op[20];
	char od[20];
	char os[20];
	int i=0,j=0;

	int v=0;	//the var num
	varNote varTable[20];	//max numbers of variables :20

	printf("enter your statememts, memory ID:%d\n",M.ID);
	printf("%-4d",j);
	scanf("%s%s%s",op,od,os);
	getchar();
	o=A.trans(op,j);  //need to change; trans char to int opcode
	if(o!=9100 && o!=1){	//not loc statement and not comment
//		d=A.trans(od,varTable,M.ptrTable[i+1],j);	//trans char to int operand
//		s=A.trans(os,varTable,M.ptrTable[i+2],j);
	}
	while(o!=9000 && o!=-1 && i<M.mSize){	//900000:end, -1:syntax error
		if(o==9100){ //note new var
			if((od[0] >='a' && od[0] <='z') || (od[0] >='A' && od[0] <='Z')){	//var name must begin with letter
				strcpy(varTable[v].var,od);	//note the var name
				varTable[v].pos=atoi(os);	//note the var pos
				varTable[v].valid=1;		//valid==1 means the position in varTable has been used
				v++;
			}
			else{
				
				RES=13;
				if(systemChecker.showLevel(RES, sysLog)){
					printf("D%-5d variable name:\"%s\" ",ID,od);	//var name must begin with letter
					systemChecker.check(RES, sysLog);
				}
				return -1;
			}
		}
		else{	//loc statement with line num 0
			if(o==1){
				;	//comment, do nothing
			}
			else{
				M.write(i,o);	//write to MCode
				M.write(i+1,d);
				M.write(i+2,s);
				i+=3;
				j++;	//LOC statement not increase the num of line
			}
		}
		printf("%-4d",j);	//num of line
		scanf("%s%s%s",op,od,os);
		getchar();
		o=A.trans(op,j);
		if(o!=9100 && o!=1){
//			d=A.trans(od,varTable,M.ptrTable[i+1],j);
//			s=A.trans(os,varTable,M.ptrTable[i+2],j);
		}		
	}
	if(o==9000){	//normal end
		printf("end edit\n");
		M.write(i,o);
		return 0;
	}
	else{	//abend
		if(o==-1){	
			
			RES=14;
			if(systemChecker.showLevel(RES, sysLog)){
				printf("D%-5d opcode:\"%s\" ",ID,op);
				systemChecker.check(RES, sysLog);
			}
		}
		else{
			
			RES=15;
			if(systemChecker.showLevel(RES, sysLog)){
				printf("D%-5d numbers of lines:%d ",ID,i/3);
				systemChecker.check(RES, sysLog);
			}
		}
		return -1;
	}
}
*/

/*
int editor::editorCodeFromFile(FILE *pfile, memory &M, assembler &A){	//editor in line
	int o,d,s;	//opcode, destination operand and source oprand
	char line[5];
	char op[20];
	char od[20];
	char os[20];
	int opcodeType=0; //with 0,1 or 2 operand
	int addressing=0; //address visit way
	int addressingUnion;	//used to handle double operand addressing
	int i=0,j=0;

	int v=0;	//the var num
	varNote varTable[20];	//max numbers of variables :20

	if(displayMode) printf("display your statememts, memory ID:%d\n",M.ID);
	if(displayMode) printf("%-4d",j);
	fscanf(pfile,"%s%s%s%s",line,op,od,os);
	if(displayMode) printf("%s %s %s\n",op,od,os);
	//getchar();

	o=A.trans(op,j);  //need to change; trans char to int opcode

	opcodeType=o%10;

	if(o!=910002 && o!=1 && opcodeType!=0 ){	//not loc statement and not comment
		d=A.trans2(od,varTable,j,addressing);	//trans char to int operand
		addressingUnion=addressing;
		s=A.trans2(os,varTable,j,addressing);
		addressingUnion+=addressing*1000;
	}
	while(o!=900000 && o!=-1 && i<M.mSize){	//900000:end, -1:syntax error
		if(o==910002){ //note new var
			if((od[0] >='a' && od[0] <='z') || (od[0] >='A' && od[0] <='Z')){	//var name must begin with letter
				strcpy(varTable[v].var,od);	//note the var name
				varTable[v].pos=atoi(os);	//note the var pos
				varTable[v].valid=1;		//valid==1 means the position in varTable has been used
				v++;
			}
			else{
				
				RES=13;
				if(systemChecker.showLevel(RES, sysLog)){
					printf("D%-5d variable name:\"%s\" ",ID,od);	//var name must begin with letter
					systemChecker.check(RES, sysLog);
				}
				return -1;
			}
		}
		else{	//loc statement with line num 0
			if(o==1){
				;	//comment, do nothing
			}
			else{
				M.write(i,o);	//write to MCode
				M.write(i+1,addressingUnion);	//visit address way
				M.write(i+2,d);
				M.write(i+3,s);
				i+=4;
				j++;	//LOC statement not increase the num of line
			}
		}
		if(displayMode) printf("%-4d",j);	//num of line
		fscanf(pfile,"%s%s%s%s",line,op,od,os);
		if(displayMode) printf("%s %s %s\n",op,od,os);
		//getchar();

		o=A.trans(op,j);

		opcodeType=o%10;

		if(o!=910002 && o!=1 && opcodeType!=0 ){
			d=A.trans2(od,varTable,j,addressing);
			addressingUnion=addressing;
			s=A.trans2(os,varTable,j,addressing);
			addressingUnion+=addressing*1000;

		//	printf("%d",addressingUnion);
		}		
	}
	if(o==900000){	//normal end
		if(displayMode) printf("end edit\n");
		M.write(i,o);
		return 0;
	}
	else{	//abend
		if(o==-1){	
			
			RES=14;
			if(systemChecker.showLevel(RES, sysLog)){
				printf("D%-5d opcode:\"%s\" ",ID,op);
				systemChecker.check(RES, sysLog);
			}
		}
		else{
			
			RES=15;
			if(systemChecker.showLevel(RES, sysLog)){
				printf("D%-5d numbers of lines:%d ",ID,i/4);
				systemChecker.check(RES, sysLog);
			}
		}
		return -1;
	}
}
*/
/*
int editor::editorCodeToFile(FILE *pfile, memory &M, assembler &A){	//editor in line
	int o,d,s;	//opcode, destination operand and source oprand
	char op[20];
	char od[20];
	char os[20];
	int i=0,j=0;

	int v=0;	//the var num
	varNote varTable[20];	//max numbers of variables :20

	printf("enter your statememts, memory ID:%d\n",M.ID);
	printf("%-4d",j);
	scanf("%s%s%s",op,od,os);
	fprintf(pfile,"%-4d%s %s %s\n",j,op,od,os);
	getchar();
	o=A.trans(op,j);  //need to change; trans char to int opcode
	if(o!=9100 && o!=1){	//not loc statement and not comment
//		d=A.trans(od,varTable,M.ptrTable[i+1],j);	//trans char to int operand
//		s=A.trans(os,varTable,M.ptrTable[i+2],j);
	}
	while(o!=9000 && o!=-1 && i<M.mSize){	//9000:end, -1:syntax error
		if(o==9100){ //note new var
			if((od[0] >='a' && od[0] <='z') || (od[0] >='A' && od[0] <='Z')){	//var name must begin with letter
				strcpy(varTable[v].var,od);	//note the var name
				varTable[v].pos=atoi(os);	//note the var pos
				varTable[v].valid=1;		//valid==1 means the position in varTable has been used
				v++;
			}
			else{
				
				RES=13;
				if(systemChecker.showLevel(RES, sysLog)){
					printf("D%-5d variable name:\"%s\" ",ID,od);	//var name must begin with letter
					systemChecker.check(RES, sysLog);		
				}
				return -1;
			}
		}
		else{	//loc statement with line num 0
			if(o==1){
				;	//comment, do nothing
			}
			else{
				M.write(i,o);	//write to MCode
				M.write(i+1,d);
				M.write(i+2,s);
				i+=3;
				j++;	//LOC statement not increase the num of line
			}
		}
		printf("%-4d",j);	//num of line
		scanf("%s%s%s",op,od,os);
		fprintf(pfile,"%-4d%s %s %s\n",j,op,od,os);
		getchar();
		o=A.trans(op,j);
		if(o!=9100 && o!=1){
//			d=A.trans(od,varTable,M.ptrTable[i+1],j);
//			s=A.trans(os,varTable,M.ptrTable[i+2],j);
		}		
	}
	if(o==9000){	//normal end
		printf("end edit\n");
		M.write(i,o);
		return 0;
	}
	else{	//abend
		if(o==-1){	
			
			RES=14;
			if(systemChecker.showLevel(RES, sysLog)){
				printf("D%-5d opcode:\"%s\" ",ID,op);
				systemChecker.check(RES, sysLog);	
			}
		}
		else{
			
			RES=15;
			if(systemChecker.showLevel(RES, sysLog)){
				printf("D%-5d numbers of lines:%d ",ID,i/3);
				systemChecker.check(RES, sysLog);
			}
		}
		return -1;
	}
}
*/

int interpreter::effectAddressing(int addcode, int fadd,process &proc, int &RorM){
	int R;
	int type;
	int EA;

	R=addcode/100;	//偏移寻址时使用的通用寄存器号
	type=addcode%100;	//寻址方式
	RorM=0;	//0，操作数在主存；1，操作数在寄存器；2，操作数为立即数。0 memory;1 register; 2 immediate

	switch(type){//具体有效地址的计算方法详见文档《有效地址的计算》
		case 0: EA=fadd;
				break;
		case 2: EA=fadd;
				RorM=1;
				break;
		case 3: EA=fadd;
				RorM=2;
				break;
		case 4: EA=proc.MData.read(fadd);
				break;
		case 6: EA=GM.read(fadd);
				break;
		case 8:	EA=GM.read(R)+fadd;
				break;
		case 12:EA=GM.read(R)+proc.MData.read(fadd);
				break;
		case 14:EA=GM.read(R)+GM.read(fadd);
				break;
		default: printf("error!");
				return -1;
	}

	return EA;	//返回有效地址


}

int interpreter::exer(process &proc){	//MCode:Code memory, MData:Data memory, m: temp memory
	//int codeChangeFlag[2]={0,0};
	//C.CM=0;	//note the position of units in MCode and m
	int i=0;	//用于计数程序执行的次数。
	int ppk;	//保存通用寄存器组的值时，用于计数

	if(proc.exetime==0){//第一次执行
		PC=0; //first execute
		proc.exetime=1;
	}	
	else{
		pop(flag,proc.S);
		pop(PC,proc.S);
		for(ppk=14; ppk>=0; ppk--)
			pop(GM.mem[ppk],proc.S);
		//C.pop(C.flag,proc.S);
		//C.pop(C.CM,proc.S);
	}


	load(MRgst,0,proc.MCode,PC);	PC++;	//load opcode
	proc.exetime++;	//exetime add 1
	
	while(i<cycleTimes && MRgst.read(0)!=900000 && MRgst.read(0)!=800000 && PC<proc.MCode.mSize){ //8000 meams HALT 
		
	/*
 		if(proc.MCode.ptrTable[C.CM+1]==1){
			codeChangeFlag[0]=1;	//indicate the code unit has been modified
		    C.mov(MRgst,3,proc.MCode,C.CM+1);	//protect code to m
			C.mov(proc.MCode,C.CM+1,proc.MData,proc.MCode.read(C.CM+1));
			//modify code unit: use the ptr var's value(in MData) to substitute its address(in MCode) 
			//ERROR! MCode.ptrTable[C.CM]=0;	//erase ptr flag
		}
		if(proc.MCode.ptrTable[C.CM+2]==1){	
			codeChangeFlag[1]=1;	//indicate the code unit has been modified
			C.mov(MRgst,4,proc.MCode,C.CM+2);	//protect code to m
			C.mov(proc.MCode,C.CM+2,proc.MData,proc.MCode.read(C.CM+2));
			//ERROR! MCode.ptrTable[C.CM+1]=0;	//erase ptr flag
		}
	*/	

		int addVisit_s, addVisit_d;	//源操作数和目的操作数的寻址方式
		int opType;	//操作数类型
		int EA_s, EA_d;	//源操作数和目的操作数的有效地址。effective address
		int RorM_s,RorM_d;//操作数存放的位置
		int dataLS;//数据通路信息，详见文档《操作码的数据操作方式》
		
		
		opType=MRgst.read(0)%10;//操作码的最后一位表示操作数的个数
		dataLS=(MRgst.read(0)%100)/10;//操作码的倒数第二位表示操作码的数据操作方式

		load(MRgst,1,proc.MCode,PC);	PC++;	//load addressing way
		load(MRgst,2,proc.MCode,PC);	PC++;	//load operand s
		load(MRgst,4,proc.MCode,PC);	PC++;	//load operand d
		
		addVisit_s=MRgst.read(1)%1000;	//first addressing
		addVisit_d=MRgst.read(1)/1000;	//second addressing

		if(opType==0){//无操作数
			MRgst.write(3,-1);	//mean no operation。用以标记误操作数
		}
		else
			if(opType==1){//单操作数
				EA_s=effectAddressing(addVisit_s,MRgst.read(2),proc,RorM_s);	//计算有效地址
				MRgst.write(2,EA_s);	//保存有效地址到内部寄存器
				MRgst.write(3,RorM_s);	//保存操作数存放的位置
			}
			else
				if(opType==2){//双操作数
					EA_s=effectAddressing(addVisit_s,MRgst.read(2),proc,RorM_s);
					MRgst.write(2,EA_s);
					MRgst.write(3,RorM_s);
	
					EA_d=effectAddressing(addVisit_d,MRgst.read(4),proc,RorM_d);
					MRgst.write(4,EA_d);
					MRgst.write(5,RorM_d);
				}

		int tmp_dataLS=dataLS;	//（临时）操作码的数据操作方式

		if(tmp_dataLS%2 == 1 && MRgst.read(5)==0){
			load(GM,16,proc.MData,MRgst.read(4));	//需要将源操作数的数值保存到数据寄存器（位于通用寄存器组的16位）
		}
		
		tmp_dataLS>>=1;
	
		if(tmp_dataLS%2 == 1 && MRgst.read(3)==0){
			load(GM,15,proc.MData,MRgst.read(2));	//需要将目的操作数的数值保存到数据寄存器（位于通用寄存器组的15位）
		}

		tmp_dataLS>>=1;

		if(MRgst.read(3)==1) GM.mem[15]=GM.mem[MRgst.read(2)];	//operand in reg (GM);
		if(MRgst.read(5)==1) GM.mem[16]=GM.mem[MRgst.read(4)];
		if(MRgst.read(3)==2) GM.mem[15]=MRgst.read(2);	//immediate;
		if(MRgst.read(5)==2) GM.mem[16]=MRgst.read(4);

		switch(MRgst.read(0)){   
			case 900152: //C.mov(proc.MRgst,1,proc.MCode,C.CM); C.CM++;	//mov; place the 1st operand to m
					//C.mov(proc.MRgst,2,proc.MCode,C.CM); C.CM++;		//place the 2nd operand to m
					//C.mov(proc.MData,MRgst.read(1),MRgst.read(2));	//exeucute and mov ans to MData
					GM.mem[15]=GM.mem[16];
					break;
			//case 90112: 
			case 902042: //C.mov(proc.MRgst,1,proc.MCode,C.CM); C.CM++;   //(lea)
					//C.mov(proc.MRgst,2,proc.MCode,C.CM); C.CM++;
					GM.write(15,MRgst.read(4));	//RorM_d cant be 1!!!
					break;
			case 900272: //C.mov(proc.MRgst,1,proc.MCode,C.CM); C.CM++;	//add
					//C.mov(proc.MRgst,2,proc.MCode,C.CM); C.CM++;
					//C.add(proc.MData,MRgst.read(1),MRgst.read(2));
					GM.mem[15]+=GM.mem[16];
					break;
			case 901072: //C.mov(proc.MRgst,1,proc.MCode,C.CM); C.CM++;	//sub
					//C.mov(proc.MRgst,2,proc.MCode,C.CM); C.CM++;
					//C.sub(proc.MData,MRgst.read(1),MRgst.read(2));
					GM.mem[15]-=GM.mem[16];
					break;
			case 902072:												//mul
					GM.mem[15]*=GM.mem[16];
					break;
			case 903072:												//div
					int tmp;
					tmp=GM.mem[15];
					GM.mem[15]/=GM.mem[16];
					GM.mem[0]=tmp % GM.mem[16]; //保存余数
					break;
			case 904072:												//mod
					GM.mem[15]%=GM.mem[16];
					break;
			case 905072:												//and
					GM.mem[15]&=GM.mem[16];
					break;
			case 906072:												//or
					GM.mem[15]|=GM.mem[16];
					break;
			case 907072:												//xor
					GM.mem[15]^=GM.mem[16];
					break;
			//case 90122: //C.mov(proc.MRgst,1,proc.MCode,C.CM); C.CM++;   //Imadd
			//		//C.mov(proc.MRgst,2,proc.MCode,C.CM); C.CM++;
			//		C.Imadd(proc.MData,MRgst.read(1),MRgst.read(2));
			//		break;
			//case 90142:// C.mov(proc.MRgst,1,proc.MCode,C.CM); C.CM++;   //Imsub
			//		//C.mov(proc.MRgst,2,proc.MCode,C.CM); C.CM++;
			//		C.Imsub(proc.MData,MRgst.read(1),MRgst.read(2));
			//		break;
			case 900332:// C.mov(proc.MRgst,1,proc.MCode,C.CM); C.CM++;	//cmp
					//C.mov(proc.MRgst,2,proc.MCode,C.CM); C.CM++;
					//C.cmp(proc.MData,MRgst.read(1),MRgst.read(2)); 
					int r;
					if((r=GM.mem[15]-GM.mem[16])>0) flag=1;
					else
						if(r<0) flag=-1;
						else flag=0;
					break;
			//case 9013: //C.mov(proc.MRgst,1,proc.MCode,C.CM); C.CM++;	//Imcmp
			//		//C.mov(proc.MRgst,2,proc.MCode,C.CM); C.CM++;
			//		C.Imcmp(proc.MData,MRgst.read(1),MRgst.read(2)); 
			//		break;
			case 900421: //C.mov(proc.MRgst,1,proc.MCode,C.CM); C.CM++;   //ja
					//C.CM++;	//the opcode has only one operand, so there is a blank line need to cross
					//if(C.flag>0) C.ja(MRgst.read(1));
					//if(flag>0) { PC=4*GM.mem[15]; flag=0;}
					if(flag>0) { PC=GM.mem[15]; flag=0;}	//在新的跳转语句中，直接跳到对应行号，不需要再乘以4
					break;
			case 900521: //C.mov(proc.MRgst,1,proc.MCode,C.CM); C.CM++;	//jb
					//C.CM++;
					//if(C.flag<0) C.jb(MRgst.read(1));
					if(flag<0) { PC=GM.mem[15]; flag=0;}
					break;
			case 900621: //C.mov(proc.MRgst,1,proc.MCode,C.CM); C.CM++;	//jmp
					//C.CM++;
					//C.jmp(MRgst.read(1));
					PC=GM.mem[15]; flag=0;
					break;
			case 900721: //C.mov(proc.MRgst,1,proc.MCode,C.CM); C.CM++;	//je
					//C.CM++;
					//if(C.flag==0) C.je(MRgst.read(1));
					if(flag==0) { PC=GM.mem[15]; flag=0;}
					break;
			case 900821: //C.mov(proc.MRgst,1,proc.MCode,C.CM); C.CM++;	//jne
					//C.CM++;
					//if(C.flag!=0) C.jne(MRgst.read(1));
					if(flag!=0) { PC=GM.mem[15]; flag=0;}
					break;
			case 900961: //C.mov(proc.MRgst,1,proc.MCode,C.CM); C.CM++;	//inc
					// C.CM++;
					//C.inc(proc.MData,MRgst.read(1));
					GM.mem[15]++;
					break;
			case 901961:												//neg
					GM.mem[15]=-GM.mem[15];
					break;
			case 902961:												//not
					GM.mem[15]=~GM.mem[15];
					break;
			case 902121: //C.mov(proc.MRgst,1,proc.MCode,C.CM); C.CM++;	//push
					//C.CM++;
					//C.push(proc.MData,MRgst.read(1),proc.S);
					push(GM.mem[15],proc.S);
					break;
			case 902300:												//pusha
					for(ppk=0; ppk<15; ppk++)
						push(GM.mem[ppk],proc.S);
					break;
			case 902241:// C.mov(proc.MRgst,1,proc.MCode,C.CM); C.CM++;	//pop
					//C.CM++;
					//C.pop(proc.MData,MRgst.read(1),proc.S);
					pop(GM.mem[15],proc.S);
					break;
			case 902400:												//popa
					for(ppk=14; ppk>=0; ppk--)
						pop(GM.mem[ppk],proc.S);
					break;
			//case 90302: //C.mov(proc.MRgst,1,proc.MCode,C.CM); C.CM++;	//ofst
			//		//C.mov(proc.MRgst,2,proc.MCode,C.CM); C.CM++;
			//		C.offset(proc.MData,MRgst.read(1),MRgst.read(2));
			//		break;
			case 904000:// C.mov(proc.MRgst,1,proc.MCode,C.CM); C.CM++;	//int
					//C.CM++;
					//C.push(C.CM,proc.S);
					//C.push(C.flag,proc.S);
					for(ppk=0; ppk<15; ppk++)	//中断，保存现场信息
						push(GM.mem[ppk],proc.S);
					push(PC,proc.S);
					push(flag,proc.S);
					
					RES=20;	
					if(systemChecker.showLevel(RES, sysLog)){
						printf("D%-5d process ID:%d, instruction row:%d ", ID,proc.ID, PC/4-1);
						systemChecker.check(RES, sysLog);	//interrupt happened
					}
					return GM.mem[10];	//int return
					break;
			case 904121: //C.mov(proc.MRgst,1,proc.MCode,C.CM); C.CM++; //call
					//C.CM++;
					//C.push(C.CM,proc.S);
					//C.push(C.flag,proc.S);
					push(PC,proc.S);
					push(flag,proc.S);

					//C.mov(MRgst,1,proc.MData,MRgst.read(1)); //change the fun's address in MData to its value 
					//C.CM=4*MRgst.read(1);	//!!! violate rule ,directly from mem!!!
					PC=GM.mem[15]; flag=0;	//改为相同
					//PC=4*GM.mem[15];  //和跳转语句不同
					break;
			case 904200: //C.CM+=2;										//ret
					//C.pop(C.flag,proc.S);
					//C.pop(C.CM,proc.S);
					pop(flag,proc.S);
					pop(PC,proc.S);
					break;
			case 904300: //C.mov(proc.MRgst,1,proc.MCode,C.CM); C.CM++;	//iret
					//C.CM++;
					return 3; //return from syscall
					break;
			case 950000:// C.mov(proc.MRgst,1,proc.MCode,C.CM); C.CM++;	//sysr
					//C.mov(proc.MRgst,2,proc.MCode,C.CM); C.CM++;
					//C.mov(MRgst,1,proc.MData,MRgst.read(1));
					//C.mov(MRgst,2,proc.MData,MRgst.read(2));
					//C.push(C.CM,proc.S);
					//C.push(C.flag,proc.S);

					for(ppk=0; ppk<15; ppk++)
						push(GM.mem[ppk],proc.S);
					push(PC,proc.S);
					push(flag,proc.S);

					return 5;	//int return
					break;
			case 950100:// C.mov(proc.MRgst,1,proc.MCode,C.CM); C.CM++;	//sysw
					//C.mov(proc.MRgst,2,proc.MCode,C.CM); C.CM++;
					//C.mov(MRgst,1,proc.MData,MRgst.read(1));
					//C.mov(MRgst,2,proc.MData,MRgst.read(2));
					//C.push(C.CM,proc.S);
					//C.push(C.flag,proc.S);
					
					for(ppk=0; ppk<15; ppk++)
						push(GM.mem[ppk],proc.S);
					push(PC,proc.S);
					push(flag,proc.S);

					return 6;	//int return
					break;
			case 950200: //C.mov(proc.MRgst,1,proc.MCode,C.CM); C.CM++;	//wake
					//C.mov(proc.MRgst,2,proc.MCode,C.CM); C.CM++;
					//C.mov(MRgst,1,proc.MData,MRgst.read(1));
					//C.mov(MRgst,2,proc.MData,MRgst.read(2));
					//C.push(C.CM,proc.S);
					//C.push(C.flag,proc.S);

					for(ppk=0; ppk<15; ppk++)
						push(GM.mem[ppk],proc.S);
					push(PC,proc.S);
					push(flag,proc.S);

					return 7;	//int return
					break;
			case 950300:// C.mov(proc.MRgst,1,proc.MCode,C.CM); C.CM++;	//set
					//C.mov(proc.MRgst,2,proc.MCode,C.CM); C.CM++;
					switch(GM.read(11)){	//maybe its also a violation
						case 0: proc.communicate=GM.read(12);break;	
						default: 	
								RES=24;
								if(systemChecker.showLevel(RES, sysLog)){
									printf("D%-5d process ID:%d, property NO:%d ",ID,proc.ID,GM.read(11));	//illegal property NO.
									systemChecker.check(RES, sysLog);
								}
					}
					proc.update=1;
					break;
			case 960000:// C.CM+=2;										//cli
					proc.interrupt=1;
					proc.update=1;
					break;
			case 960100: //C.CM+=2;										//sti
					proc.interrupt=0;
					proc.update=1;
					break;


			//case 91002: C.CM+=2;										//loc
					//break;
			default: ;
		}

		/*
		if(codeChangeFlag[0]==1){
			C.mov(proc.MCode,C.CM-2,MRgst,3);	//rewrite code
			codeChangeFlag[0]=0;
		}
		if(codeChangeFlag[1]==1){
			C.mov(proc.MCode,C.CM-1,MRgst,4);	//rewrite code
			codeChangeFlag[1]=0;
		}
		*/

		if(tmp_dataLS==1 && MRgst.read(3)==0){
			store(proc.MData,MRgst.read(2),GM,15);	//将运算结果写回主存
		}

		if(tmp_dataLS==1 && MRgst.read(3)==1) GM.mem[MRgst.read(2)]=GM.mem[15];	//运算结果写回寄存器。operand in reg (GM);

		load(MRgst,0,proc.MCode,PC);	PC++;	//load next opcode
		proc.exetime++;	//exetime add 1
		i++;
	}

	if(i==cycleTimes){	//解释器周期到（解释器的周期在建立解释器对象时设定），进程每一轮运行周期包含的解释器周期数用进程的优先级规定
		PC--;

		for(ppk=0; ppk<15; ppk++)	//保存现场
			push(GM.mem[ppk],proc.S);
		push(PC,proc.S);
		push(flag,proc.S);

		//C.push(C.CM,proc.S);
		//C.push(C.flag,proc.S); //I ignore this!
		
		RES=22;
		if(systemChecker.showLevel(RES, sysLog)){
			printf("D%-5d process ID:%d ",ID,proc.ID);
			systemChecker.check(RES, sysLog);
		}
		return 1;
	}
	else	//解释器溢出
		if(PC==proc.MCode.mSize){	//this condition maybe not happen //happen!!!
			
			RES=21;
			if(systemChecker.showLevel(RES, sysLog)){
				printf("D%-5d process ID:%d ",ID,proc.ID);
				systemChecker.check(RES, sysLog);
			}
			return -1;
		}
		else{ //END 900000 or HALT 800000。进程运行结束
			
			RES=23;
			if(systemChecker.showLevel(RES, sysLog)){
				printf("D%-5d process ID:%d ",ID,proc.ID);
				systemChecker.check(RES, sysLog);
			}
			return 0;
		}
}

/*
int userInterface::menu(processor &P1, interpreter &I1, process &proc, checker &CK, editor &e, assembler &a){
	char c;	//pressed key
	char filename[30];
	int size; //memory size
	//char fileflag=0; //whether the file has been open
	printf("           XmipsUSER1.0\n");
	printf("------- C'Upper Pseudo ASM -------\n");	//title
	printf("e--edit new file\nf--load existed file\nm--show memory\nu--update memory\ni--system info\na--about\nq--quit\n");	//menu
	c=getch();	//get your choice
	while(c!='q' && c!='Q'){	//q to quit
		if(c=='e' || c=='E'){	//edit code
			printf("\nsave(y/n):");
			c=getch();
			while(c!='y' && c!='Y' && c!='n' && c!='N') c=getch();
			if(c=='y' || c=='Y'){
				printf("new file name:");	//create a new and save
				gets(filename);
				if((f1=fopen(filename,"w"))==NULL){
					printf("D%-5d filename:\"%s\" ",ID,filename);	//fail to open
					RES=2;
					systemChecker.check(RES, sysLog);
					getch();
				}
				else{
				//	printf("\nedit code:\n");
				//	if(!e.editorCodeFromFile(proc.MCode,f1,a))
				//		I1.exer(proc);
				//	else
				//		printf("not execute\n");
					fclose(f1);
					getch();
				}
			}
			else{	//just edit not save
				printf("\nedit code:\n");
//				if(!e.editorCollection(proc.MCode,1,a))
//					I1.exer(proc);
//				else
//					printf("not execute\n");
				getch();
			}
		}
		else
			if(c=='f' || c=='F'){	//load code
				printf("\ninput file name: ");
				gets(filename);
				if((f1=fopen(filename,"r"))==NULL){
					printf("D%-5d filename:\"%s\" ",ID,filename);;	//fail to open
					RES=1;
					systemChecker.check(RES, sysLog);
					getch();
				}
				else{
				//	if(!e.editorCollection(proc.MCode,f1,0,a))
				//		I1.exer(proc);
				//	else
				//		printf("not execute\n");
					fclose(f1);
					getch();
				}
			}
			else
				if(c=='m' || c=='M'){	//show memory
					printf("\nchoose memory:\nc--code\nd--data\nr--register\ns--stack\na--all memories\n");
					c=getch();
					switch(c){
						case 'c':
						case 'C':printf("Code Memory:\n");showMemory(proc.MCode);break;
						case 'd':
						case 'D':printf("Data Memory:\n");showMemory(proc.MData);break;
		//				case 'r':
		//				case 'R':printf("Register Memory:\n");showMemory(proc.MRgst);break;
						case 's':
						case 'S':printf("Stack Memory:\n");showMemory(proc.S);break;
						case 'a':
						case 'A':printf("Code Memory:\n");showMemory(proc.MCode);
								printf("Data Memory:\n");showMemory(proc.MData);
//								printf("Register Memory:\n");showMemory(proc.MRgst);
								printf("Stack:\n");showMemory(proc.S);break;
						default: ;
					}
					getch();
				}
				else{
					if(c=='u' || c=='U'){
						printf("\nmemory size:\nc--code (%d)\nd--data (%d)\ns--stack (%d)\n",
							proc.MCode.mSize,proc.MData.mSize,proc.S.mSize);
						c=getch();
						switch(c){
							case 'c':
							case 'C':printf("new code memory size:");
								scanf("%d",&size);
								getchar();
								proc.MCode.update(size);
								break;
							case 'd':
							case 'D':printf("new data memory size:");
								scanf("%d",&size); 	
								proc.MData.update(size);
								break;
//							case 'r':
//							case 'R':printf("new register memory size:");
//								scanf("%d",&size); 	
//								proc.MRgst.update(size);
//								break;
							case 's':
							case 'S':printf("new stack size:");
								scanf("%d",&size); 	
								proc.S.update(size);
								break;
							default: ;
						}
						printf("memory updated\n");
						getch();
					}
					else
						if(c=='i' || c=='I'){
							printf("\nDevice		Type		ID\n");
							printf("MCode		memory		%d\n", proc.MCode.ID);
							printf("MData		memory		%d\n", proc.MData.ID);
//							printf("MRgst		memory		%d\n", proc.MRgst.ID);
							printf("S1		stack		%d\n", proc.S.ID);
							printf("P1		processor	%d\n", P1.ID);
							printf("I1		interpretor	%d\n", I1.ID);
							getch();
						}
						else
							if(c=='a' || c=='A'){
								printf("\nsystem: XmipsUSER1.0\n");
								printf("language: C'Upper Pseudo ASM\n");
								printf("Copyright (C) Yu Chenye 2011\n");
								getch();
							}
			}

		printf("\n           XmipsUSER1.0\n");
		printf("------- C'Upper Pseudo ASM -------\n");	//title
		printf("e--edit new file\nf--load existed file\nm--show memory\nu--update memory\ni--system info\na--about\nq--quit\n");	//menu
		c=getch();
	}
	return 0;
}
*/
/*	
int userInterface::showMemory(memory &M){
	int i;
	printf("memory ID:%d\n",M.ID);
	for(i=0; i<M.mSize; i++)
		printf("%-4d%d\n",i,M.read(i)); //show M.mem[i]
	return 0;
}
*/
queue::enQueue(procptr pptr){
	if((tail+1) % length == head){	//queue is full
	
		RES=60;
		if(systemChecker.showLevel(RES, sysLog)){
			printf("D%-5d process ID:%d queue size:%d member's number:%d ",ID,pptr->ID,length-1,memberNumber);
			systemChecker.check(RES, sysLog);
		}
		return -1;
	}
	else{
		procQueue[tail]=pptr;
		tail=(tail+1) % length;
		memberNumber++;
		return 0;
	}
}

queue::deQueue(procptr &pptr){
	if(head==tail){	//queue is empty
		
		RES=61;
		if(systemChecker.showLevel(RES, sysLog)){
			printf("D%-5d ",ID);
			systemChecker.check(RES, sysLog);
		}
		return -1;
	}
	else{
		pptr=procQueue[head];
		head=(head+1) % length;
		memberNumber--;
		return 0;
	}
}

int list::put(int t, procptr pptr){
	if(t<0 || t>=length){
		return -1;	//over range
	}
	else{
		listp[t]=pptr;
		return 0;
	}//put will cover the unit whether it has process ptr or not
}

int list::get(int t, procptr &pptr){
	if(t<0 || t>=length){
		return -1;	//over range
	}
	else
		if(listp[t]==0){
			return -1; //proc not exist
		}
		else{
			pptr=listp[t];
			return 0;
		}
}

int sysList::copy(int t, procptr &dpptr){
	if(t<0 || t>=length){
		return -1;	//over range
	}
	else
		if(listp[t]==0){
			return -1; //proc not exist
		}
		else{
			dpptr=new process(t,1,*listp[t]); //lost t before!!! notice 1 (2nd parameter)
			
			RES=90;
			if(systemChecker.showLevel(RES, sysLog)){
				printf("D%-5d system function ID:%d ",ID,t);
				systemChecker.check(RES, sysLog);
			}
			return 0;
		}
}
	
int sysList::copy(procptr &dpptr, procptr spptr){ //dont use this copy
	dpptr=new process(0,1,*spptr); //copy (*spptr) to (*dpptr)
	return 0;
}	

int pcbList::insertToHead(pcbptr pblock){	//insert a pcb to the head of a list
	pblock->next=head->next;
	head->next=pblock;
	len++;
	return 0;
}

int pcbList::enQueue(pcbptr pblock){	//insert the block to the tail of the list by refer its ptr
	tail->next=pblock;	//tail->next pointer points to current added pcb node
	tail=tail->next;
	pblock->next=0;	//the end of the list is null
	len++;
	return 0;
}

int pcbList::deQueue(pcbptr &pblock){	//return the ptr of the first block which is deleted from the list
	if(len==0){	//the list is empty;		
	
		RES=130;
		if(systemChecker.showLevel(RES, sysLog)){
			printf("D%-5d pcb's process ID:%d ",ID,pblock->ID);
			systemChecker.check(RES, sysLog);
		}

		return -1;
	}
	pblock=head->next;	//delete the pcb next to the blank pcb
	head->next=head->next->next;
	if(len==1) tail=head;	//delete the last node, tail need to change 
	len--;
	return 0;
}

	
int pcbList::IDGet(pcbptr &pblock,int id){
	if(len==0){	//the list is empty;
		
		RES=131;
		if(systemChecker.showLevel(RES, sysLog)){
			printf("D%-5d pcb's process ID:%d ",ID,pblock->ID);
			systemChecker.check(RES, sysLog);
		}
		return -1;
	}
	pcbptr current=head->next;	//point to current pcb
	pcbptr pri=head;	//point to prior pcb
	while(current->ID!=id && current!=0){
		pri=current;
		current=current->next;
	}
	if(current==0){	//not find the pcb whose id is same to requested id
		printf("not find pcb!\n"); 
		return -1; 
	}
	else{ 
		pblock=current;
		//block.link(*(current->pptr));
		pri->next=current->next;
		if(current==tail) tail=pri;
		len--;
		return 0;
	}
}

int pcbList::CommunicateGet(pcbptr &pblock,int com){
	if(len==0){	//the list is empty;
		
		RES=131;
		if(systemChecker.showLevel(RES, sysLog)){
			printf("D%-5d ",ID);
			systemChecker.check(RES, sysLog);
		}
		return -1;
	}
	pcbptr current=head->next;	//point to current pcb
	pcbptr pri=head;	//point to prior pcb
	while(current!=0 && current->communicate!=com){ //notice the sequence, if current==0, then current->communicate is access violation
		pri=current;
		current=current->next;
	}
	if(current==0){	//not find the pcb whose id is same to requested id
		
		RES=132;
		if(systemChecker.showLevel(RES, sysLog)){
			printf("D%-5d searched pcb communicate:%d ",ID,com); 
			systemChecker.check(RES, sysLog);
		}
		return -1;
	}
	else{ 
		pblock=current;
		//block.link(*(current->pptr));
		pri->next=current->next;
		if(current==tail) tail=pri;
		len--;
		return 0;
	}
}

void pcb::copy(pcb &block){	//copy pcb to current pcb
	ID=block.pptr->ID;
	state=block.pptr->ID;
	priority=block.pptr->priority;
	exetime=block.pptr->exetime;
	callerID=block.pptr->callerID;
	communicate=block.pptr->communicate;
	interrupt=block.pptr->interrupt;

	pptr=block.pptr;
	next=0;
}

void pcb::link(process &proc){	//link a process to a pcb and use the properties of the process to change the pcb
	ID=proc.ID;
	state=proc.ID;
	priority=proc.priority;
	exetime=proc.exetime;
	callerID=proc.callerID;
	communicate=proc.communicate;
	interrupt=proc.interrupt;

	pptr=&proc;
	next=0;
}

void pcb::release(){	//return a pcb to system
	ID=state=priority=callerID=communicate=interrupt=0;
	valid=0;
	pptr=0;
	next=0;
}

void pcb::update(procptr pptr){ //update a pcb when some properties of its process is changed
	ID=pptr->ID;
	state=pptr->ID;
	priority=pptr->priority;
	exetime=pptr->exetime;
	callerID=pptr->callerID;
	communicate=pptr->communicate;
	interrupt=pptr->interrupt;
}


int storage::getFile(FILE* &pf, char *filename, int choice){
	char tmp[40];
	strcpy(tmp,pos);
	strcat(tmp,filename);
	switch(choice){	//choose file open way
		case 0: strcpy(visitWay, "r");break;
		case 1: strcpy(visitWay, "w");break;
		case 2: strcpy(visitWay, "r+");break;
		default: printf("D%-5d visit way:%d ",ID,choice);RES=80;//visit way error
			systemChecker.check(RES, sysLog);return -1;
	}
	if((pf=fopen(tmp,visitWay))==NULL){
		
		RES=81; //open failed
		if(systemChecker.showLevel(RES, sysLog)){	
			printf("D%-5d file path:\"%s\" ",ID,tmp);
			systemChecker.check(RES, sysLog);
		}
		return -1;
	}
	return 0;
}	

int storage::releaseFile(FILE* &fp){	//close file
	if(fp!=0){
		fclose(fp);
		fp=0;
		return 0;
	}
	else{
		
		RES=82;
		if(systemChecker.showLevel(RES, sysLog)){	
			printf("D%-5d ",ID);	//file pointer is null
			systemChecker.check(RES, sysLog);
		}
		return -1;
	}
}

int checker::showLevel(int k,abendNote *abendTable){	//whether show the report or not	
	int i=0;
	int rl;	//report lecel
	
	while(k!=abendTable[i].abendCode) i++;	
	
	rl=abendTable[i].level;
	if(reportLevel==0 && rl!=2 && rl!=3) return 0;	//dont show
	if(reportLevel==1 && rl==1) return 0;	//dont show
	return 1;

}

void checker::check(int k,abendNote *abendTable){	//check the RES
	int i=0;

	char *level[]={	//condition levels
		"Normal",
		"Normal(Sub)",
		"Warning",
		"Error"
	};

	while(k!=abendTable[i].abendCode) i++;
	
	printf(" %s, RES=%d, %s\n",abendTable[i].abendReason,abendTable[i].abendCode,level[abendTable[i].level]);
}
