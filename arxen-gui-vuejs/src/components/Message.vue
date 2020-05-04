<template>
    <div>
        <div class="card-date" v-if="showDate" v-html="message.timeStamp"></div>
        <div
                class="message-box"
                :class="{ 'offset-current': message.user === currentUserId }"
        >
            <div
                    class="message-container"
                    :class="{
					'message-container-offset': messageOffset
				}"
            >
                <div
                        ref="imageRef"
                        class="message-card"
                        :class="{
						'message-highlight': isMessageHover(message),
						'message-current': message.userId === currentUserId,
						'message-deleted': false
					}"
                >
                    <div
                            v-if="message.chatId"
                            class="text-username"
                            :class="{
							'username-reply': false
						}"
                    >
                        <span v-html="message.user"></span>
                    </div>
                    <span v-html="message.text"></span>

                    <div class="text-timestamp">
                        <span v-html="message.timeStamp"></span>
                    </div>
                </div>
            </div>
        </div>

    </div>
</template>

<script>
    export default {
        name: "Message",
        // tmp solution
        data() {
            return {
                currentUserId: "tcp://10.6.0.3:7878"
            }
        },
        props: {
            showDate: {type: Boolean, default: false},
            message: {
                type: Object,
            },
        },
        methods: {
            isMessageHover() {
                return false
                // this.editedMessage._id === this.message._id ||
                // this.hoverMessageId === this.message._id
            },
        },
        computed: {
            messageOffset() {
                return (
                    this.index > 0 &&
                    this.message.userId !== this.messages[this.index - 1].userId
                )
            },
        }
    };
    //<strong>{{message.user}}</strong>: {{message.text}}
</script>

<style lang="scss" scoped>
    .card-date {
        border-radius: 4px;
        max-width: 150px;
        text-align: center;
        margin: 10px auto;
        font-size: 12px;
        text-transform: uppercase;
        padding: 4px;
        color: var(--chat-message-color-date);
        background: var(--chat-message-bg-color-date);
        display: block;
        overflow-wrap: break-word;
        position: relative;
        white-space: normal;
        box-shadow: 0 1px 1px -1px rgba(0, 0, 0, 0.1),
        0 1px 1px -1px rgba(0, 0, 0, 0.11), 0 1px 2px -1px rgba(0, 0, 0, 0.11);
    }

    .line-new {
        color: var(--chat-message-color-new-messages);
        position: relative;
        text-align: center;
        font-size: 13px;
        padding: 10px 0;
    }

    .line-new:after,
    .line-new:before {
        border-top: 1px solid var(--chat-message-color-new-messages);
        content: '';
        left: 0;
        position: absolute;
        top: 50%;
        width: calc(50% - 60px);
    }

    .line-new:before {
        left: auto;
        right: 0;
    }

    .message-box {
        display: flex;
        flex: 0 0 50%;
        max-width: 50%;
        justify-content: flex-start;
        line-height: 1.4;
    }

    .message-container {
        position: relative;
        padding: 2px 10px;
        align-items: end;
        min-width: 100px;
        box-sizing: content-box;
    }

    .message-container-offset {
        margin-top: 10px;
    }

    .offset-current {
        margin-left: 50%;
        justify-content: flex-end;
    }

    .message-card {
        background: var(--chat-message-bg-color);
        color: var(--chat-message-color);
        border-radius: 8px;
        font-size: 14px;
        padding: 6px 9px 3px;
        white-space: pre-wrap;
        max-width: 100%;
        -webkit-transition-property: box-shadow, opacity;
        transition-property: box-shadow, opacity;
        transition: box-shadow 280ms cubic-bezier(0.4, 0, 0.2, 1);
        will-change: box-shadow;
        box-shadow: 0 1px 1px -1px rgba(0, 0, 0, 0.1),
        0 1px 1px -1px rgba(0, 0, 0, 0.11), 0 1px 2px -1px rgba(0, 0, 0, 0.11);
    }

    .message-highlight {
        box-shadow: 0 1px 2px -1px rgba(0, 0, 0, 0.1),
        0 1px 2px -1px rgba(0, 0, 0, 0.11), 0 1px 5px -1px rgba(0, 0, 0, 0.11);
    }

    .message-current {
        background: var(--chat-message-bg-color-me) !important;
    }

    .message-deleted {
        color: var(--chat-message-color-deleted) !important;
        font-size: 13px !important;
        font-style: italic !important;
        background: var(--chat-message-bg-color-deleted) !important;
    }

    .image-container {
        width: 250px;
        max-width: 100%;
    }

    .image-reply-container {
        width: 70px;
    }

    .message-image {
        position: relative;
        background-color: var(--chat-message-bg-color-image) !important;
        background-size: cover !important;
        background-position: center center !important;
        background-repeat: no-repeat !important;
        height: 250px;
        width: 250px;
        max-width: 100%;
        border-radius: 4px;
        margin: 4px auto 5px;
        transition: 0.4s filter linear;
    }

    .message-image-reply {
        height: 70px;
        width: 70px;
        margin: 4px auto 3px;
    }

    .image-loading {
        filter: blur(3px);
    }

    .reply-message {
        background: var(--chat-message-bg-color-reply);
        border-radius: 4px;
        margin: -1px -5px 8px;
        padding: 8px 10px;

        .reply-username {
            color: var(--chat-message-color-reply-username);
            font-size: 12px;
            line-height: 15px;
            margin-bottom: 2px;
        }

        .reply-content {
            font-size: 12px;
            color: var(--chat-message-color-reply-content);
        }
    }

    .text-username {
        font-size: 13px;
        color: var(--chat-message-color-username);
        margin-bottom: 2px;
    }

    .username-reply {
        margin-bottom: 5px;
    }

    .text-timestamp {
        font-size: 10px;
        color: var(--chat-message-color-timestamp);
        text-align: right;
    }

    .file-message {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        margin-top: 3px;

        span {
            max-width: 100%;
        }

        .icon-file svg {
            margin-right: 5px;
        }
    }

    .options-container {
        position: absolute;
        top: 2px;
        right: 10px;
        height: 40px;
        width: 70px;
        overflow: hidden;
        z-index: 1;
        border-top-right-radius: 8px;
    }

    .blur-container {
        position: absolute;
        height: 100%;
        width: 100%;
        left: 8px;
        bottom: 10px;
        background: var(--chat-message-bg-color);
        filter: blur(3px);
        border-bottom-left-radius: 8px;
    }

    .options-me {
        background: var(--chat-message-bg-color-me);
    }

    .options-image .blur-container {
        background: rgba(255, 255, 255, 0.6);
        border-bottom-left-radius: 15px;
    }

    .image-buttons {
        position: absolute;
        width: 100%;
        height: 100%;
        border-radius: 4px;
        background: linear-gradient(
                        to bottom,
                        rgba(0, 0, 0, 0) 55%,
                        rgba(0, 0, 0, 0.02) 60%,
                        rgba(0, 0, 0, 0.05) 65%,
                        rgba(0, 0, 0, 0.1) 70%,
                        rgba(0, 0, 0, 0.2) 75%,
                        rgba(0, 0, 0, 0.3) 80%,
                        rgba(0, 0, 0, 0.5) 85%,
                        rgba(0, 0, 0, 0.6) 90%,
                        rgba(0, 0, 0, 0.7) 95%,
                        rgba(0, 0, 0, 0.8) 100%
        );

        svg {
            height: 26px;
            width: 26px;
        }

        .button-view,
        .button-download {
            position: absolute;
            bottom: 6px;
            left: 7px;
        }

        :first-child {
            left: 40px;
        }

        .button-view {
            max-width: 18px;
            bottom: 8px;
        }
    }

    .message-options {
        background: var(--chat-icon-bg-dropdown-message);
        border-radius: 50%;
        position: absolute;
        top: 7px;
        right: 7px;

        svg {
            height: 17px;
            width: 17px;
            padding: 5px;
            margin: -5px;
        }
    }

    .message-reactions {
        position: absolute;
        top: 6px;
        right: 30px;
    }

    .menu-options {
        right: 15px;
    }

    .menu-left {
        right: -118px;
    }

    .icon-check {
        height: 14px;
        width: 14px;
        vertical-align: middle;
        margin: -3px -3px 0 3px;
    }

    .button-reaction {
        display: inline-flex;
        align-items: center;
        border: var(--chat-message-border-style-reaction);
        outline: none;
        background: var(--chat-message-bg-color-reaction);
        border-radius: 4px;
        margin: 4px 2px 0;
        transition: 0.3s;
        padding: 0 5px;
        font-size: 18px;
        line-height: 23px;

        span {
            font-size: 11px;
            font-weight: 500;
            min-width: 7px;
            color: var(--chat-message-color-reaction-counter);
        }

        &:hover {
            border: var(--chat-message-border-style-reaction-hover);
            background: var(--chat-message-bg-color-reaction-hover);
            cursor: pointer;
        }
    }

    .reaction-me {
        border: var(--chat-message-border-style-reaction-me);
        background: var(--chat-message-bg-color-reaction-me);

        span {
            color: var(--chat-message-color-reaction-counter-me);
        }

        &:hover {
            border: var(--chat-message-border-style-reaction-hover-me);
            background: var(--chat-message-bg-color-reaction-hover-me);
        }
    }

    @media only screen and (max-width: 768px) {
        .message-container {
            padding: 2px 3px 1px;
        }
        .message-container-offset {
            margin-top: 10px;
        }
        .message-box {
            flex: 0 0 80%;
            max-width: 80%;
        }
        .offset-current {
            margin-left: 20%;
        }
        .options-container {
            right: 3px;
        }
        .menu-left {
            right: -50px;
        }
    }
</style>
