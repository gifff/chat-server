import json
import logging
import re
import time
import gevent
import websocket
import uuid
from locust import between, task, User
from locust_plugins.users import SocketIOUser

class WebsocketUser(User):
    abstract = True
    wait_time = between(1, 5)

    _current_msg_id = None
    _current_msg_accepted = False

    def _connect(self, host: str, header: list):
        self.ws = websocket.create_connection(host, header=header)
        gevent.spawn(self._receive_loop)

    def _receive_loop(self):
        try:
            while self.ws.connected:
                message = self.ws.recv()
                if message == '':
                    pass
                    # logging.debug(f"WebsocketUser: receive_loop: message is empty. probably a control message")
                else:
                    # logging.debug(f"WebsocketUser: receive_loop: message={message}")
                    self.on_message(message)
        except Exception as e:
            logging.exception(
                "WebsocketUser: receive_loop: fail to receiving message", exc_info=e)
            self._close()
            # raise e

    def on_message(self, message: str):
        response_time = 0
        payload = None
        try:
            payload = json.loads(message)
        except:
            logging.debug(
                f"WebsocketUser: on_message: failed to decode. message={message}")
            return

        if payload is not None and 'message' in payload:
            try:
                msg_id, sent_time_ns_str, actual_message = payload["message"].split(
                    '__')
                if msg_id != self._current_msg_id:
                    # it's receiving other message
                    return

                sent_time_ns = int(sent_time_ns_str)
                response_time = (time.time_ns() - sent_time_ns) / 1000_000.0

                self.environment.events.request.fire(
                    request_type="WSR",
                    name=actual_message,
                    response_time=response_time,
                    response_length=len(message),
                    exception=None,
                    context=self.context(),
                )

                self._current_msg_accepted = True
                # logging.debug(f"WebsocketUser: on_message: reported msg_id={msg_id} sent_time={sent_time_ns} response_time={response_time}")
            except ValueError:
                logging.error(f"WebsocketUser: on_message: invalid sent_time")
                return
            except Exception as e:
                logging.error(
                    f"WebsocketUser: on_message: invalid message. exception={e}")

    def send(self, message) -> str:
        msg_id = str(uuid.uuid4())
        current_time_ns = time.time_ns()
        payload = {
            'message': f"{msg_id}__{current_time_ns}__{message}",
            'type': 1
        }
        msg = json.dumps(payload)

        self._current_msg_accepted = False
        self._current_msg_id = msg_id

        # self.environment.events.request.fire(
        #     request_type="WSS",
        #     name=msg_id,
        #     # name=payload['message'],
        #     response_time=None,
        #     response_length=len(msg),
        #     exception=None,
        #     context=self.context()
        # )

        # logging.debug(f"WebsocketUser: send: message={msg}")
        self.ws.send(msg)
        return msg_id

    def send_and_wait(self, path: str, message: str, timeout: int = 10):
        # step 1: connect
        target_uri = self.host + path

        self._connect(target_uri, ['X-User-Id: 1000'])

        # step 2: send through ws
        self.send(message)
        sent_time = time.time()

        # step 3: wait for message to be received back OR timeout (in second)
        while True:
            if not self.ws.connected:
                break

            current_time = time.time()
            if current_time - sent_time > timeout:
                self.environment.events.request.fire(
                    request_type="WSR",
                    name=message,
                    response_time=(current_time - sent_time) * 1000,
                    response_length=len(message),
                    exception=None,
                    context=self.context(),
                )
                break
            if self._current_msg_accepted:
                break
            time.sleep(0.1)

        # step 4: close
        self._close()

    def _close(self):
        self.ws.close()


class BouncebackWebsocketUser(WebsocketUser):
    @task
    def send_and_wait_bounceback(self):
        self.send_and_wait('/messages/listen', 'anything')

    if __name__ == "__main__":
        host = "ws://localhost:9999"
