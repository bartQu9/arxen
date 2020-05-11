<template>
    <div class="col-messages" v-show="selectedChatId">
        <div class="room-header app-border-b">
            <div class="room-wrapper">
                <div
                        v-if="false"
                        class="svg-button toggle-button"
                        :class="{ 'rotate-icon': !selectedChatId }"
                >
                    <svg-icon name="toggle" />
                </div>
                <div
                        v-if="chatAvatar"
                        class="room-avatar"
                        :style="{ 'background-image': `url('${chatAvatar}')` }"
                ></div>
                <div
                        v-if="!chatAvatar"
                        class="room-avatar"
                        :style="{ 'background-image': `url('${missingAvatarUrl}')` }"
                ></div>

                <div class="room-name" v-html="selectedChatId"></div>
            </div>

            <div
                    class="svg-button room-options"
                    v-if="menuActions.length"
                    @click="menuOpened = !menuOpened"
            >
                <svg-icon name="menu"/>
            </div>
            <transition name="slide-left" v-if="menuActions.length">
                <div
                        v-if="menuOpened"
                        v-click-outside="closeMenu"
                        class="menu-options"
                >
                    <div class="menu-list">
                        <div v-for="action in menuActions" :key="action.name">
                            <div
                                    class="menu-item"
                                    v-html="action.title"
                                    @click="menuActionHandler(action)"
                            ></div>
                        </div>
                    </div>
                </div>
            </transition>
        </div>

        <div ref="scrollContainer" class="container-scroll">
            <loader :show="loadingMessages"></loader>
            <div class="messages-container">
                <div :class="{ 'messages-hidden': loadingMessages }">
                    <transition name="fade-message">
                        <div
                                class="text-started"
                                v-if="showNoMessages"
                                v-html="textMessages.MESSAGES_EMPTY"
                        ></div>
                        <div
                                class="text-started"
                                v-if="showMessagesStarted"
                                v-html="
								textMessages.CONVERSATION_STARTED + ' ' + humanDate(messages[0].timeStamp)
							"
                        ></div>
                    </transition>
                    <transition name="fade-message">
                        <infinite-loading
                                v-if="messages.length"
                                spinner="spiral"
                                direction="top"
                                @infinite="loadMoreMessages"
                        >
                            <div slot="spinner">
                                <loader :show="true" :infinite="true"></loader>
                            </div>
                            <div slot="no-results"></div>
                            <div slot="no-more"></div>
                        </infinite-loading>
                    </transition>

                    <div v-if="selectedChatId && selectedChatId.length && messages">
                        <app-message v-for="message of messages"
                                     :key="message.messageId"
                                     :message="message"
                                     :userName="userName"
                        >
                        </app-message>
                    </div>

                </div>
            </div>
        </div>

        <div v-if="!loadingMessages">
            <transition name="bounce">
                <div class="icon-scroll" v-if="scrollIcon" @click="scrollToBottom">
                    <svg-icon name="dropdown" param="scroll"/>
                </div>
            </transition>
        </div>

        <div ref="roomFooter" class="room-footer">
            <div class="box-footer">
            <textarea ref="roomTextarea"
                      :placeholder="textMessages.TYPE_MESSAGE"
                      v-model="messageInput"
                      @input="onChangeInput"
            ></textarea>

                <div class="icon-textarea">
                    <div
                            @click="onPostClick"
                            class="svg-button"
                            :class="{ 'send-disabled': inputDisabled }"
                    >
                        <svg-icon name="send" :param="inputDisabled ? 'disabled' : ''"/>
                    </div>

                </div>
            </div>

        </div>
    </div>
</template>

<script>
    import gql from 'graphql-tag';
    import Message from '@/components/Message';
    import Loader from "@/components/Loader";
    import InfiniteLoading from 'vue-infinite-loading'
    import SvgIcon from "@/components/SvgIcon";
    import vClickOutside from 'v-click-outside'
    import {missingAvatarUrl} from '@/themes';
    import {parseJSON, formatRelative} from 'date-fns'

    export default {
        name: "MessageList",
        components: {
            'app-message': Message,
            'loader': Loader,
            'infinite-loading': InfiniteLoading,
            'svg-icon': SvgIcon,
        },
        directives: {
            clickOutside: vClickOutside.directive
        },
        props: {
            selectedChatId: {type: String},
            messagesLoaded: {type: Boolean, default: true},
            textMessages: {type: Object, default: null},
            userName: {type: String},
            latestSeenMessageID: {type: String},
        },
        data() {
            return {
                currentlySelectedChat: null,
                chats: [],
                loadingMessages: false,
                loadingMoreMessages: false,
                infiniteState: null,
                scrollIcon: false,
                messages: [],
                messageInput: '',
                missingAvatarUrl,
                chatAvatar: '',
                menuOpened: false,
                menuActions: [{title: "Add", name: "add"}, {title: "Show Users", name: "show_user"}],
            };
        },
        apollo: {
            messages() {
                return {
                    query: gql`
                    query($chatID: String!) {
                            messages(chatID: $chatID) {
                                chatId
                                user
                                text
                                timeStamp
                                messageId
                            }
                    }
                    `,
                    variables() {
                        return {chatID: this.selectedChatId}
                    },
                    subscribeToMore: {
                        document: gql`
                        subscription($chatID: String!) {
                            messagePosted(chatID: $chatID) {
                                chatId
                                user
                                text
                                timeStamp
                            }
                        }`,
                        variables() {
                            return {chatID: this.selectedChatId}
                        },
                        updateQuery: (prev, {subscriptionData}) => {
                            if (!subscriptionData.data) {
                                return prev;
                            }
                            const message = subscriptionData.data.messagePosted;
                            if (prev.messages.find((m) => m === message)) {
                                return prev;
                            }
                            return Object.assign({}, prev, {
                                messages: [message, ...prev.messages],
                            });
                        },
                    },
                };
            },
        },
        watch: {
            loadingMessages(val) {
                if (val) this.infiniteState = null;
                else this.focusTextarea(true)
            },
            messagesLoaded(val) {
                if (val) this.loadingMessages = false;
                if (this.infiniteState) this.infiniteState.complete()
            },
        },
        created() {
            // tmp solution
            this.$apollo
                .query({
                    query: gql`{ chats { chatId clientsIPsList chatName } }`,
                })
                .then((data) => {
                    console.log(data);
                    this.chats = data.data.chats;
                    this.currentlySelectedChat = this.chats.find(chat => chat.chatId === this.selectedChatId);
                })
                .catch((e) => {
                    console.error(e);
                });
        },
        methods: {
            loadMoreMessages($state) {
                // phantom solution
                if (this.$apollo.loading) {
                    return
                } else {
                    return $state.complete();
                }
                // if (this.loadingMoreMessages) return
                // if (this.messagesLoaded || !this.room.roomId) {
                //     return infiniteState.complete()
                // }
                // this.infiniteState = infiniteState
                // //this.$emit('fetchMessages')
                // this.loadingMoreMessages = true
            },
            onPostClick() {
                const messageInput = this.messageInput;
                this.$apollo
                    .mutate({
                        mutation: gql`mutation($chatID: String!, $text: String!) {postMessage(chatID: $chatID, text: $text) { chatId user text timeStamp }}`,
                        variables: {
                            chatID: this.selectedChatId,
                            text: messageInput,
                        },
                    })
                    .then(() => {
                        this.messageInput = '';
                    })
                    .catch((e) => {
                        console.error(e);
                    });
            },
            resizeTextarea() {
                const el = this.$refs['roomTextarea'];
                const padding = window
                    .getComputedStyle(el, null)
                    .getPropertyValue('padding-top')
                    .replace('px', '');
                el.style.height = 0;
                el.style.height = el.scrollHeight - padding * 2 + 'px'
            },
            isMessageEmpty() {
                return !this.messageInput.trim()
            },
            onChangeInput() {
                this.resizeTextarea();
                //this.$emit('typingMessage', this.message)
            },
            humanDate(s) {
                // in GoLang time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", metadata["timeStamp"].(string))
                //console.log(s);
                let res = parseJSON(s, '2006-01-02 15:04:05.999999999 -0700 MST', new Date());
                //console.log(res);
                return formatRelative(res, new Date());
            },
            closeMenu() {
                this.menuOpened = false
            },
            menuActionHandler(action) {
                this.closeMenu();
                this.$emit('menuActionHandler', action)
            },
        },
        computed: {
            showNoMessages() {
                return (
                    !this.messages.length
                )
            },
            showMessagesStarted() {
                return this.messages.length
            },
            inputDisabled() {
                return this.isMessageEmpty()
            },
        }
    };
    // until unique property not achieved
    //<transition-group name="fade-message">

    // <div v-if="!selectedChatId">
    //     Empty
    //     </div>
    //
    //     <div v-if="!messages.length && selectedChatId.length">No Messages</div>
</script>

<style lang="scss" scoped>
    .col-messages {
        position: relative;
        height: 100%;
        flex: 1;
        overflow: hidden;
        display: flex;
        flex-flow: column;
    }

    .room-header {
        position: absolute;
        display: flex;
        align-items: center;
        height: 64px;
        width: 100%;
        z-index: 10;
        margin-right: 1px;
        background: var(--chat-header-bg-color);
        border-top-right-radius: var(--chat-container-border-radius);
    }

    .room-wrapper {
        display: flex;
        align-items: center;
        width: 100%;
        padding: 0 16px;
    }

    .toggle-button {
        margin-right: 15px;

        svg {
            height: 26px;
            width: 26px;
        }
    }

    .rotate-icon {
        transform: rotate(180deg) !important;
    }

    .room-name {
        font-size: 17px;
        font-weight: 500;
        line-height: 22px;
        color: var(--chat-header-color-name);
    }

    .room-info {
        font-size: 13px;
        line-height: 18px;
        color: var(--chat-header-color-info);
    }

    .room-options {
        margin-left: auto;
    }

    .container-scroll {
        background: var(--chat-content-bg-color);
        flex: 1;
        overflow-y: scroll;
        margin-right: 1px;
        margin-top: 60px;
        -webkit-overflow-scrolling: touch;
    }

    .messages-container {
        padding: 0 5px 5px;
    }

    .text-started {
        font-size: 14px;
        color: var(--chat-message-color-started);
        font-style: italic;
        text-align: center;
        margin-top: 30px;
        margin-bottom: 20px;
    }

    .icon-scroll {
        position: absolute;
        bottom: 80px;
        right: 20px;
        padding: 8px;
        background: var(--chat-bg-scroll-icon);
        border-radius: 50%;
        box-shadow: 0 1px 1px -1px rgba(0, 0, 0, 0.2), 0 1px 1px 0 rgba(0, 0, 0, 0.14),
        0 1px 2px 0 rgba(0, 0, 0, 0.12);
        display: flex;
        cursor: pointer;

        svg {
            height: 25px;
            width: 25px;
        }
    }

    .room-footer {
        width: calc(100% - 1px);
        border-bottom-right-radius: 4px;
        z-index: 10;
    }

    .box-footer {
        display: flex;
        position: relative;
        background: var(--chat-footer-bg-color);
        padding: 10px 8px 10px;
    }

    .reply-container {
        display: flex;
        padding: 10px 10px 0 10px;
        background: var(--chat-content-bg-color);
        align-items: center;
        max-width: 100%;

        .reply-box {
            width: 100%;
            overflow: hidden;
            background: var(--chat-footer-bg-color-reply);
            border-radius: 4px;
            padding: 8px 10px;
            display: flex;
        }

        .reply-info {
            overflow: hidden;
        }

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

        .icon-reply {
            margin-left: 10px;

            svg {
                height: 20px;
                width: 20px;
            }
        }

        .image-reply {
            max-height: 100px;
            margin-right: 10px;
        }
    }

    textarea {
        height: 20px;
        width: 100%;
        line-height: 20px;
        overflow: hidden;
        outline: 0;
        resize: none;
        border-radius: 20px;
        padding: 12px 16px;
        box-sizing: content-box;
        font-size: 16px;
        background: var(--chat-bg-color-input);
        color: var(--chat-color);
        caret-color: var(--chat-color-caret);
        border: var(--chat-border-style-input);

        &::placeholder {
            color: var(--chat-color-placeholder);
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }
    }

    .textarea-outline {
        border: 1px solid var(--chat-border-color-input-selected);
        box-shadow: inset 0px 0px 0px 1px var(--chat-border-color-input-selected);
    }

    .icon-textarea {
        display: flex;
        margin: 12px 0 0 5px;

        svg,
        .wrapper {
            margin: 0 7px;
        }
    }

    .image-container {
        position: absolute;
        max-width: 25%;
        left: 16px;
        top: 18px;
    }

    .image-file {
        display: flex;
        justify-content: center;
        flex-direction: column;
        min-height: 30px;

        img {
            border-radius: 15px;
            width: 100%;
            max-width: 150px;
            max-height: 100%;
        }
    }

    .icon-image {
        position: absolute;
        top: 6px;
        left: 6px;
        z-index: 10;

        svg {
            height: 20px;
            width: 20px;
            border-radius: 50%;
        }

        &:before {
            content: ' ';
            position: absolute;
            width: 100%;
            height: 100%;
            background: rgba(0, 0, 0, 0.5);
            border-radius: 50%;
            z-index: -1;
        }
    }

    .file-container {
        display: flex;
        align-items: center;
        width: calc(100% - 75px);
        height: 20px;
        padding: 12px 0;
        box-sizing: content-box;
        background: var(--chat-bg-color-input);
        border: var(--chat-border-style-input);
        border-radius: 20px;
    }

    .file-container-edit {
        width: calc(100% - 109px);
    }

    .file-message {
        max-width: calc(100% - 75px);
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .icon-file {
        display: flex;
        margin: 0 8px 0 15px;
    }

    .icon-remove {
        margin-left: 8px;

        svg {
            height: 18px;
            width: 18px;
        }
    }

    .send-disabled,
    .send-disabled svg {
        cursor: none !important;
        pointer-events: none !important;
        transform: none !important;
    }

    .messages-hidden {
        opacity: 0;
    }

    @media only screen and (max-width: 768px) {
        .room-header {
            height: 50px;

            .room-wrapper {
                padding: 0 10px;
            }

            .room-name {
                font-size: 16px;
                line-height: 22px;
            }

            .room-info {
                font-size: 12px;
                line-height: 16px;
            }

            .room-avatar {
                height: 37px;
                width: 37px;
            }
        }
        .container-scroll {
            margin-top: 50px;
        }
        .box-footer {
            border-top: var(--chat-border-style-input);
            padding: 7px 2px 7px 7px;
        }
        .text-started {
            margin-top: 20px;
        }
        textarea {
            padding: 7px;
            line-height: 18px;

            &::placeholder {
                color: transparent;
            }
        }
        .icon-textarea {
            margin: 6px 0 0 5px;

            svg,
            .wrapper {
                margin: 0 5px;
            }
        }
        .image-container {
            top: 10px;
            left: 10px;
        }
        .image-file img {
            transform: scale(0.97);
        }
        .room-footer {
            width: 100%;
        }
        .file-container {
            padding: 7px 0;

            .icon-file {
                margin-left: 10px;
            }
        }
        .reply-container {
            padding: 5px 8px;
        }
        .icon-scroll {
            bottom: 70px;
        }
    }
</style>
