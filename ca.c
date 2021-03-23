#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <malloc.h>
#include <string.h>

#include <infiniband/verbs.h>

#include "ca.h"
#include "caint.h"

// cast b to a or a <= b returns c on failure
#define IBVContextCast(a,b,c)	\
		if ((b) == NULL) { \
			return (c); \
		} \
		(a) = (pIBVContext) (b);

void test_ibv_struct(IBVAddress *pIBVA)
{
	printf("lid : %d\n", pIBVA->lid);
	printf("qpn : %d\n", pIBVA->qpn);
	printf("psn : %d\n", pIBVA->psn);
	printf("raddr : %ld\n", pIBVA->raddr);
	printf("rkey : %d\n", pIBVA->rkey);
}

char *go_ibv_get_error(OPAQUE op)
{
	pIBVContext ibvc;
	IBVContextCast(ibvc,op,NULL)
	return ibvc->err_str;
}

int go_ibv_get_ibv_address_non_rdma_nogen(OPAQUE op, IBVAddress *pIBVA)
{
	pIBVContext ibvc;
	IBVContextCast(ibvc,op,-1)
	pIBVA->lid = ibvc->portattr.lid;
	pIBVA->qpn = ibvc->qp->qp_num;
	pIBVA->psn = ibvc->local_psn;
	ibvc->local_psn = pIBVA->psn;
}

// used during send/receive semantics
int go_ibv_get_ibv_address_non_rdma(OPAQUE op, IBVAddress *pIBVA)
{
	pIBVContext ibvc;
	IBVContextCast(ibvc,op,-1)
	pIBVA->lid = ibvc->portattr.lid;
	pIBVA->qpn = ibvc->qp->qp_num;
	pIBVA->psn = lrand48() & 0xFFFFF;
	ibvc->local_psn = pIBVA->psn;
}

// used during rdma semantics
int go_ibv_get_ibv_address(OPAQUE op, IBVAddress *pIBVA)
{
	pIBVContext ibvc;
	IBVContextCast(ibvc,op,-1)
	pIBVA->lid = ibvc->portattr.lid;
	pIBVA->qpn = ibvc->qp->qp_num;
	pIBVA->psn = lrand48() & 0xFFFFF;
	pIBVA->raddr = (uintptr_t) ibvc->buf;
	pIBVA->rkey = ibvc->mr->rkey;
}

// rand48 seed + any other startup stuff
void go_ibv_init()
{
	srand48(getpid() * time(NULL));
}

int go_ibv_get_device_list_num(OPAQUE op) 
{
	int num_list;
	pIBVContext ibvc;

	IBVContextCast(ibvc,op,-1)

	ibvc->mtu =IBV_MTU_2048; // should be a parameter

	return ibvc->dev_list_num;
}

OPAQUE go_ibv_get_device_list() {
	int num_list;
	pIBVContext ibvc;
	OPAQUE rval;
	ibvc = (pIBVContext) malloc(sizeof(IBVContext));
	memset(ibvc,0,sizeof(IBVContext));
	ibvc->dev_list = ibv_get_device_list(&ibvc->dev_list_num);
	printf("go_ibv_get_device_list() : %p\n", ibvc->dev_list);
	rval = (OPAQUE) ibvc;
	return rval;
}

int go_ibv_set_device_index(OPAQUE op, int index)
{
	pIBVContext ibvc;
	IBVContextCast(ibvc,op,index)
	ibvc->dev = ibvc->dev_list[index];
	ibvc->dev_index = index;
	return index;
}

int go_ibv_free_device_list(OPAQUE p)
{
	pIBVContext ibvc;
	IBVContextCast(ibvc,p,-1)
/*
	if (p == NULL)
	{
		return;
	}
	ibvc = (pIBVContext) p;
*/
	printf("go_ibv_free_device_list() : %p\n", ibvc->dev_list);
	ibv_free_device_list((struct ibv_device**) ibvc->dev_list);
	return 0;
}

const char* go_ibv_get_device_name(OPAQUE op, int index)
{
	pIBVContext ibvc;
	struct ibv_device *p;
	IBVContextCast(ibvc,op,NULL);
	if (index == -1) 
	{
		p = ibvc->dev;
	}else{
		p = ibvc->dev_list[index];
	}
printf("device name : %s\n", ibv_get_device_name(p));
	return ibv_get_device_name(p);
}

int go_ibv_open_device(OPAQUE op)
{
	pIBVContext ibvc;
	IBVContextCast(ibvc,op,-1)
	ibvc->context = ibv_open_device(ibvc->dev);
	if (ibvc->context==NULL)
	{
		ibvc->err_str = strdup("nil context");
		return -1;
	}
	return 0;
}

int go_ibv_close_device(OPAQUE op)
{
	int rval;
	pIBVContext ibvc;
	IBVContextCast(ibvc,op,-1)
	rval = ibv_close_device(ibvc->context);
	return rval;
}

void* get_aligned_buffer(unsigned int size)
{
	void *p;
	int pgsz;
	pgsz = getpagesize();
	p = (void*) memalign((size_t) pgsz, (size_t) size);
	if (p==NULL)
	{
		return NULL;
	}
	memset(p,0,size);
	return p;
}

void *go_ibv_alloc_buffer(OPAQUE op, int size)
{
	pIBVContext ibvc;
	IBVContextCast(ibvc,op,NULL);
	if (size < 1)
	{
		ibvc->buf = NULL;
		ibvc->size = 0;
		return NULL;
	}
	ibvc->buf = get_aligned_buffer((unsigned int) size);
	if (ibvc->buf == NULL) {
		ibvc->size = 0;
		return NULL;
	}
	ibvc->size = size;
	return ibvc->buf;
}

int go_ibv_create_comp_channel(OPAQUE op)
{
	pIBVContext ibvc;
	IBVContextCast(ibvc,op,-1)
	if( ibvc->use_event == 0 )
	{
		ibvc->channel = NULL;
		return 0;
	}
	ibvc->channel = ibv_create_comp_channel(ibvc->context);
	if (ibvc->channel==NULL)
	{
		ibvc->err_str = strdup("unable to create completion channel");
		return -1;
	}
	return 0;
}

int go_ibv_destroy_comp_channel(OPAQUE op)
{
	pIBVContext ibvc;
	IBVContextCast(ibvc,op,-1)
	if (ibvc->channel != NULL)
	{
		ibv_destroy_comp_channel(ibvc->channel);
	}
	return 0;
}

int go_ibv_alloc_pd(OPAQUE op)
{
	pIBVContext ibvc;
	IBVContextCast(ibvc,op,-1)
	ibvc->pd = ibv_alloc_pd(ibvc->context);
	if (ibvc->pd == NULL)
	{
		ibvc->err_str = strdup("unable to create permission domain");
		return -1;
	}
printf("pd: %p\n", ibvc->pd);
	return 0;
}

int go_ibv_dealloc_pd(OPAQUE op)
{
	pIBVContext ibvc;
	IBVContextCast(ibvc,op,-1)
	if (ibvc->pd != NULL) 
	{
		ibv_dealloc_pd(ibvc->pd);
	}
	return 0;
}


int go_ibv_create_cq(OPAQUE op, int sdepth, int rdepth)
{
	pIBVContext ibvc;
	IBVContextCast(ibvc,op,-1)
	ibvc->scq = ibv_create_cq(ibvc->context,rdepth,NULL,ibvc->channel,0);
	if (ibvc->scq == NULL) 
	{
		ibvc->err_str = strdup("unable to create cq");
		return -1;
	}

// temporarily use a single cq ??
/*
	ibvc->rcq = ibv_create_cq(ibvc->context,sdepth,NULL,ibvc->channel,0);
	if (ibvc->scq == NULL) 
	{
		ibvc->err_str = strdup("unable to create cq");
		return -1;
	}
*/
	ibvc->cq = ibvc->rcq = ibvc->scq;
	ibvc->rx_depth = rdepth;
	ibvc->sx_depth = sdepth;
printf("sx : %p rx : %p\n", ibvc->scq, ibvc->rcq);
	return 0;
}

int go_ibv_destroy_cq(OPAQUE op)
{
	pIBVContext ibvc;
	IBVContextCast(ibvc,op,-1)
	if (ibvc->scq != NULL)
	{
		ibv_destroy_cq(ibvc->scq);
	}
	if (ibvc->rcq != NULL)
	{
		ibv_destroy_cq(ibvc->rcq);
	}
	return 0;
}

// should allow more options
int go_ibv_simple_create_qp(OPAQUE op)
{
	pIBVContext ibvc;
// putting this on a stack SIGSEGV
	struct ibv_qp_init_attr *attr;
	IBVContextCast(ibvc,op,-1)

	attr = (struct ibv_qp_init_attr*) malloc(sizeof(struct ibv_qp_init_attr));
	memset(attr, 0, sizeof(struct ibv_qp_init_attr));

	attr->send_cq = ibvc->scq;
	attr->recv_cq = ibvc->rcq;
	attr->cap.max_send_wr = ibvc->sx_depth;
	attr->cap.max_recv_wr = ibvc->rx_depth;
	attr->cap.max_send_sge = 1;
	attr->cap.max_recv_sge = 1;
	attr->qp_type = IBV_QPT_RC; // reliable connection
	ibvc->qp = ibv_create_qp(ibvc->pd,attr);
	if (ibvc->qp == NULL)
	{
		ibvc->err_str = strdup("unable to create queue pair");
		return -1;
	}
	return 0;
}

int go_ibv_init_qp(OPAQUE op, unsigned char port)
{
	pIBVContext ibvc;
	struct ibv_qp_attr *qp_attr;
	IBVContextCast(ibvc,op,-1)

	ibvc->port = port;

	qp_attr = (struct ibv_qp_attr*) malloc(sizeof(struct ibv_qp_attr));

	qp_attr->qp_state = IBV_QPS_INIT;
	qp_attr->pkey_index = 0;
	qp_attr->port_num = port; // ???
	qp_attr->qp_access_flags = 0;

	ibv_modify_qp(ibvc->qp, qp_attr,
					IBV_QP_STATE | IBV_QP_PKEY_INDEX | IBV_QP_PORT | IBV_QP_ACCESS_FLAGS);

	
	return 0;
}

int go_ibv_query_port(OPAQUE op, unsigned char port)
{
	int rval;
	pIBVContext ibvc;
	IBVContextCast(ibvc, op, -1);
	rval = ibv_query_port(ibvc->context, port, &ibvc->portattr);
	if (rval==0)
	{
		printf("lid : %d\n", ibvc->portattr.lid);
	}
	if (rval != 0)
	{
		printf("ibv_query_port() error : %s\n", strerror(errno));
	}
	return rval;
}

int go_ibv_reg_mr(OPAQUE op)
{
	pIBVContext ibvc;
	IBVContextCast(ibvc,op,-1)
	ibvc->mr = ibv_reg_mr(ibvc->pd,ibvc->buf,ibvc->size,IBV_ACCESS_LOCAL_WRITE);
	if (ibvc->mr==NULL)
	{
		ibvc->err_str = strdup("unable to register memory");
		return -1;
	}
	return 0;
}



void* get_buf(unsigned int sz)
{
	void* p;
	CA *pca;
	int pgsz;
	pgsz = getpagesize();
	printf("pagesize : %d\n", pgsz);
	p = (void*) memalign((size_t) pgsz, (size_t) sz);
	if (p == NULL)
	{
		return NULL;
	}
	memset(p,0,sz);
	pca = (CA*) p;
	pca->i = 0;
	pca->j = 1;
	pca->k = 2;
	printf("ca : %p\n", p);
	return p;
}

//
// not sure if this is required on both ends ??
//
int go_ibv_post_recv(OPAQUE op, int recv_wr_id)
{
	pIBVContext ibvc;
	struct ibv_sge list;
	struct ibv_recv_wr wr;

	IBVContextCast(ibvc,op,-1)
	

	ibvc->recv_wr_id = recv_wr_id;
	list.addr = (uintptr_t) ibvc->buf;

	list.length = ibvc->size;
	list.lkey = ibvc->mr->lkey;

	wr.wr_id = recv_wr_id;
	wr.sg_list = &list;
	wr.num_sge = 1;
	struct ibv_recv_wr *bad_wr;
	int i;
	for(i=0;i<ibvc->rx_depth;i++)
	{
		if (ibv_post_recv(ibvc->qp, &wr, &bad_wr))
			break;
	}
	return i; // number of recvs posted
}

int go_ibv_query_gid(OPAQUE op, int port, int index, IBVGid *ibv_gid)
{
	pIBVContext ibvc;
	IBVContextCast(ibvc,op,-1)
	union ibv_gid gid;
	if (ibv_query_gid(ibvc->context, port, index, &gid))
	{
		ibv_gid->subnet_prefix = 0;
		ibv_gid->interface_id = 0;
		return -1;
	}
	ibv_gid->subnet_prefix = gid.global.subnet_prefix;
	ibv_gid->interface_id = gid.global.interface_id;
	return 0;
}

int go_ibv_modify_qp_remote_ep(OPAQUE op, uint32 qpn, uint32 psn, uint16 dlid)
{
	int sl = 0; // hard code service level to 0 - not sure what this means
	pIBVContext ibvc;
	IBVContextCast(ibvc,op,-1)
	ibvc->remote_psn = psn;
	struct ibv_qp_attr attr = {
		.qp_state = IBV_QPS_RTR,
		.path_mtu = ibvc->mtu, // fix
		.dest_qp_num = qpn,
		.rq_psn = psn,
		.max_dest_rd_atomic = 1,
		.min_rnr_timer = 12,
		.ah_attr = {
			.is_global = 0,
			.dlid = dlid,
			.sl = sl, // fix
			.src_path_bits = 0,
			.port_num = ibvc->port // fix
		}
	};
	if (ibv_modify_qp(ibvc->qp, &attr,
						IBV_QP_STATE | IBV_QP_AV | IBV_QP_PATH_MTU | IBV_QP_DEST_QPN |
						IBV_QP_RQ_PSN | IBV_QP_MAX_DEST_RD_ATOMIC | IBV_QP_MIN_RNR_TIMER)) {
			ibvc->err_str = strdup("Failed to modify qp to RTR");
			return -1;
	}

	attr.qp_state = IBV_QPS_RTS;
	attr.timeout = 14;
	attr.retry_cnt = 7; // infinite retries
	attr.rnr_retry = 7;
	attr.sq_psn = ibvc->local_psn; // my psn
	attr.max_rd_atomic = 1;
	if ( ibv_modify_qp(ibvc->qp, &attr,
			IBV_QP_STATE | IBV_QP_TIMEOUT | IBV_QP_RETRY_CNT |
			IBV_QP_RNR_RETRY | IBV_QP_SQ_PSN | IBV_QP_MAX_QP_RD_ATOMIC)) {
		ibvc->err_str = strdup("Failed to modify QP to RTS");
		return -1;
	}
	
	return 0;
}

void display_x(char * xs, int n)
{
	int i;
	fprintf(stderr, "Value: ");
	for(i=0;i<n;i++)
	{
		fprintf(stderr, "%02x", xs[i]);
	}
	fprintf(stderr, "\n");
}

int go_ibv_post_send(OPAQUE op, int send_wr_id, int size)
{
	pIBVContext ibvc;
	IBVContextCast(ibvc,op,-1)

// these structs may SEGV
fprintf(stderr, "Sending : %d bytes\n", size); // ibvc->size);
display_x(ibvc->buf, 200);
	ibvc->send_wr_id = send_wr_id;
	struct ibv_sge list = {
		.addr = (uintptr_t) ibvc->buf,
		.length = size, // ibvc->size,
		.lkey = ibvc->mr->lkey
	};
	struct ibv_send_wr wr = {
		.wr_id = send_wr_id,
		.sg_list = &list,
		.num_sge = 1,
		.opcode = IBV_WR_SEND,
		.send_flags = IBV_SEND_SIGNALED	
	};
	struct ibv_send_wr *bad_wr;
	return ibv_post_send(ibvc->qp, &wr, &bad_wr);
}

int go_ibv_toggle_use_event(OPAQUE op)
{
	pIBVContext ibvc;
	IBVContextCast(ibvc,op,-1)

	if ( ibvc->use_event == 0 )
	{
		ibvc->use_event = 1;
	}else{
		ibvc->use_event = 0;
	}
	return ibvc->use_event;
}

int go_ibv_req_notify_cq(OPAQUE op)
{
	pIBVContext ibvc;
	IBVContextCast(ibvc,op,-1)
	if ( ibvc->use_event == 0 )
	{
		return 0; // do not request notifications
	}
	if ( ibv_req_notify_cq(ibvc->cq, 0) )
	{
		ibvc->err_str = strdup("Couldn't request CQ notofication");
		return -1;
	}
	return 0;
}

int go_ibv_recv_req_notify_cq(OPAQUE op)
{
	pIBVContext ibvc;
	IBVContextCast(ibvc,op,-1)
	if ( ibv_req_notify_cq(ibvc->rcq, 0) )
	{
		ibvc->err_str = strdup("Couldn't request CQ notofication");
		return -1;
	}
	return 0;
}

int go_ibv_send_req_notify_cq(OPAQUE op)
{
	pIBVContext ibvc;
	IBVContextCast(ibvc,op,-1)
	if ( ibv_req_notify_cq(ibvc->scq, 0) )
	{
		ibvc->err_str = strdup("Couldn't request CQ notofication");
		return -1;
	}
	return 0;
}

int go_ibv_poll_cq_event(OPAQUE op, int use_event)
{
	pIBVContext ibvc;
	void *ev_ctx;
	struct ibv_cq *ev_cq;
	int num_cq_events = 0; // may move to our communication area
	
	IBVContextCast(ibvc,op,-1)

	if(ibvc->use_event == 1)
	{

// wait for a cq event

printf("Waiting for cq event\n");
		if ( ibv_get_cq_event(ibvc->channel, &ev_cq, &ev_ctx)) {
			ibvc->err_str = strdup("Failed to get cq_event");
			return -1;	
		}
printf("Recevied %d events\n", num_cq_events);
		num_cq_events++;
	
		if (ev_cq != ibvc->cq)
		{
			ibvc->err_str = strdup("CQ event for unknown CQ!!");
			return -1;
		}

// repost desire for another event notification

		if ( ibv_req_notify_cq(ibvc->cq, 0) ) {
			ibvc->err_str = strdup("couldn't request CQ notification");
			return -1;
		}
	}

	struct ibv_wc wc[2];
	memset(wc, 0, 2 * sizeof(struct ibv_wc));
	int ne,i;

// poll for a cq event
// this will spin if events are not used - timeout/countout
	do
	{
		ne = ibv_poll_cq(ibvc->cq, 2, wc);
		if ( ne < 0 )
		{
			ibvc->err_str = strdup("poll CQ failed!");
			return -1;
		}
	}while( !ibvc->use_event && ne < 1);

	int false;
	ibvc->scnt=0;
	ibvc->rcnt=0;
	for (i = 0; i < ne; i++)
	{
		if(wc[i].status != IBV_WC_SUCCESS)
		{
			ibvc->err_str = (char*) malloc(1024);
			sprintf(ibvc->err_str, "Failed status %s (%d) for %d",
							ibv_wc_status_str(wc[i].status),
							wc[i].status,
							(int) wc[i].wr_id);
			return -1;
		}

// switch statement
		if ( (int) wc[i].wr_id == ibvc->recv_wr_id )
		{
			ibvc->rcnt++;
		fprintf(stderr ,"recv : %d bytes\n", wc[i].byte_len);
			goto cont;
		}
		if ( (int) wc[i].wr_id == ibvc->send_wr_id )
		{
// test our receive counts and post more recv entries
		fprintf(stderr ,"sent : %d bytes status : %d\n", wc[i].byte_len,wc[i].status);
			ibvc->scnt++;
			goto cont;
		}
// default
cont:
		false=0; // stop compiler complaining ???
	}

// if we have used events
// ack our events to ensure that destroying the cq will not hang
// ack multiple to avoid multiple mutex locks
	if ( ibvc->use_event == 1 ) // maybe we should always call this
	{
		ibv_ack_cq_events(ibvc->cq, num_cq_events);
	}


	return 0;
}

int go_ibv_cleanup(OPAQUE op)
{
	pIBVContext ibvc;
	IBVContextCast(ibvc, op, -1)
	if (ibvc->qp)
	{
		ibv_destroy_qp(ibvc->qp);
		ibvc->qp = NULL;
	}
	if (ibvc->cq)
	{
		ibv_destroy_cq(ibvc->cq);
		ibvc->qp = NULL;
	}
	if (ibvc->mr)
	{
		ibv_dereg_mr(ibvc->mr);
		ibvc->mr = NULL;
	}
	if (ibvc->pd)
	{
		ibv_dealloc_pd( ibvc->pd );
		ibvc->pd = NULL;
	}
	if (ibvc->channel)
	{
		ibv_destroy_comp_channel(ibvc->channel);
		ibvc->channel = NULL;
	}
	if (ibvc->context)
	{
		ibv_close_device( ibvc->context );
		ibvc->context = NULL;
	}
	free( ibvc->buf );
	ibvc->buf = NULL;
	return 0;
}

