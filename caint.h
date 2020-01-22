#ifndef CAINT_H
#define CAINT_H

// our unholy monolith structure

typedef struct {
	int dev_list_num;
	int dev_index;
	struct ibv_device ** dev_list;
	struct ibv_device *dev;
	struct ibv_context *context;
	struct ibv_comp_channel *channel;	
	struct ibv_pd *pd; // protection domain
	struct ibv_mr *mr; // memory region
	struct ibv_cq *cq; // send completion queue
	struct ibv_cq *scq; // send completion queue
	struct ibv_cq *rcq; // recv completion queue
	struct ibv_qp *qp; // queue pair
	
	void *buf; // memory buffer
	int size; // size of memory buf
	int rx_depth; // request depth
	int sx_depth; // send depth
	int port;
	int pending;  // used for polling on completion queue
	int mtu;
	int local_psn; // my psn value
	int remote_psn; // remote end psn

	int use_event; // flag for using events
	int recv_wr_id; // our recv id
	int send_wr_id; // out send id
	int scnt;
	int rcnt;

	struct ibv_port_attr	portattr; // channel port

	char *err_str; // library provided error string
	int err_no; // ibv_ provided error number
} IBVContext;

typedef IBVContext *pIBVContext;

#endif /* CAINT_H */


