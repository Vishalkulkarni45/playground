'use client';

import React, { useEffect, useRef, useState } from 'react';
import type { SelfApp } from '@selfxyz/common';
import { QRCodeSVG } from 'qrcode.react';
import { v4 as uuidv4 } from 'uuid';
import { getUniversalLink, REDIRECT_URL, WS_DB_RELAYER } from '@selfxyz/common';
import { io } from 'socket.io-client';

// Angular QR code steps (copied from Angular SDK)
const QRcodeSteps = {
    DISCONNECTED: 0,
    WAITING_FOR_MOBILE: 1,
    MOBILE_CONNECTED: 2,
    PROOF_GENERATION_STARTED: 3,
    PROOF_GENERATION_FAILED: 4,
    PROOF_GENERATED: 5,
    PROOF_VERIFIED: 6,
};

// Angular-style WebSocket implementation (copied from Angular SDK)
const initWebSocket = (
    websocketUrl: string,
    selfApp: SelfApp,
    type: 'websocket' | 'deeplink',
    setProofStep: (step: number) => void,
    onSuccess: () => void,
    onError: (data: { error_code?: string; reason?: string }) => void
) => {
    const validateWebSocketUrl = (url: string) => {
        if (url.includes('localhost') || url.includes('127.0.0.1')) {
            throw new Error('localhost websocket URLs are not allowed');
        }
    };

    validateWebSocketUrl(websocketUrl);
    const sessionId = selfApp.sessionId;
    console.log(`[Angular QR] Initializing WebSocket connection for sessionId: ${sessionId}`);

    const fullUrl = `${websocketUrl}/websocket`;
    const socket = io(fullUrl, {
        path: '/',
        query: { sessionId, clientType: 'web' },
        transports: ['websocket'],
    });

    const handleWebSocketMessage = async (data: { status: string; error_code?: string; reason?: string }) => {
        console.log('[Angular QR] Received mobile status:', data.status, 'for session:', sessionId);
        switch (data.status) {
            case 'mobile_connected':
                console.log('[Angular QR] Mobile device connected. Emitting self_app event');
                setProofStep(QRcodeSteps.MOBILE_CONNECTED);
                if (type === 'websocket') {
                    socket.emit('self_app', { ...selfApp, sessionId });
                }
                break;
            case 'mobile_disconnected':
                console.log('[Angular QR] Mobile device disconnected.');
                setProofStep(QRcodeSteps.WAITING_FOR_MOBILE);
                break;
            case 'proof_generation_started':
                console.log('[Angular QR] Proof generation started.');
                setProofStep(QRcodeSteps.PROOF_GENERATION_STARTED);
                break;
            case 'proof_generated':
                console.log('[Angular QR] Proof generated.');
                setProofStep(QRcodeSteps.PROOF_GENERATED);
                break;
            case 'proof_generation_failed':
                console.log('[Angular QR] Proof generation failed.');
                setProofStep(QRcodeSteps.PROOF_GENERATION_FAILED);
                onError(data);
                break;
            case 'proof_verified':
                console.log('[Angular QR] Proof verified.');
                setProofStep(QRcodeSteps.PROOF_VERIFIED);
                onSuccess();
                break;
            default:
                console.log('[Angular QR] Unhandled mobile status:', data.status);
                break;
        }
    };

    socket.on('connect', () => {
        console.log(`[Angular QR] Connected with id: ${socket.id}, transport: ${socket.io.engine.transport.name}`);
    });

    socket.on('connect_error', (error) => {
        console.error('[Angular QR] Connection error:', error);
    });

    socket.on('mobile_status', handleWebSocketMessage);

    socket.on('disconnect', (reason: string) => {
        console.log(`[Angular QR] Disconnected. Reason: ${reason}`);
    });

    return () => {
        console.log(`[Angular QR] Cleaning up connection for sessionId: ${sessionId}`);
        if (socket) {
            socket.disconnect();
        }
    };
};

interface AngularQRWrapperProps {
    selfApp: SelfApp;
    onSuccess: () => void;
    onError: (data: { error_code?: string; reason?: string }) => void;
    type?: 'websocket' | 'deeplink';
    websocketUrl?: string;
    size?: number;
    darkMode?: boolean;
}

const AngularQRWrapper: React.FC<AngularQRWrapperProps> = ({
    selfApp,
    onSuccess,
    onError,
    type = 'websocket',
    websocketUrl = WS_DB_RELAYER,
    size = 300,
    darkMode = false,
}) => {
    const [proofStep, setProofStep] = useState(QRcodeSteps.WAITING_FOR_MOBILE);
    const [sessionId, setSessionId] = useState('');
    const [qrValue, setQrValue] = useState('');
    const socketRef = useRef<ReturnType<typeof initWebSocket> | null>(null);

    // Use session ID from selfApp if provided, otherwise generate one
    useEffect(() => {
        if (selfApp.sessionId) {
            setSessionId(selfApp.sessionId);
        } else {
            const newSessionId = uuidv4();
            setSessionId(newSessionId);
        }
    }, [selfApp.sessionId]);

    // Update QR value when sessionId or type changes
    useEffect(() => {
        if (sessionId) {
            if (type === 'websocket') {
                setQrValue(`${REDIRECT_URL}?sessionId=${sessionId}`);
            } else {
                setQrValue(getUniversalLink({
                    ...selfApp,
                    sessionId: sessionId,
                }));
            }
        }
    }, [sessionId, type, selfApp]);

    // Initialize WebSocket connection
    useEffect(() => {
        if (sessionId && !socketRef.current) {
            console.log('[Angular QR] Initializing WebSocket connection');
            try {
                socketRef.current = initWebSocket(
                    websocketUrl,
                    {
                        ...selfApp,
                        sessionId: sessionId,
                    },
                    type,
                    setProofStep,
                    onSuccess,
                    onError
                );
            } catch (error) {
                console.error('[Angular QR] WebSocket initialization failed:', error);
            }
        }

        return () => {
            if (socketRef.current) {
                console.log('[Angular QR] Cleaning up WebSocket connection');
                socketRef.current();
                socketRef.current = null;
            }
        };
    }, [sessionId, type, websocketUrl, onSuccess, onError, selfApp]);

    // LED indicator colors
    const getLedColor = () => {
        if (proofStep >= QRcodeSteps.MOBILE_CONNECTED) {
            return '#31F040'; // Green
        } else if (proofStep >= QRcodeSteps.WAITING_FOR_MOBILE) {
            return '#424AD8'; // Blue
        } else {
            return '#95a5a6'; // Gray
        }
    };

    // Render different states
    const renderContent = () => {
        switch (proofStep) {
            case QRcodeSteps.PROOF_GENERATION_STARTED:
            case QRcodeSteps.PROOF_GENERATED:
                return (
                    <div className="flex flex-col items-center justify-center" style={{ width: size, height: size }}>
                        <div className="animate-spin rounded-full h-16 w-16 border-4 border-green-400 border-t-transparent mb-4"></div>
                        <div className="text-green-600 text-sm">Generating Proof...</div>
                    </div>
                );

            case QRcodeSteps.PROOF_GENERATION_FAILED:
                return (
                    <div className="flex flex-col items-center justify-center" style={{ width: size, height: size }}>
                        <div className="text-red-500 text-6xl mb-4">✗</div>
                        <div className="text-red-600 text-sm">Verification Failed</div>
                    </div>
                );

            case QRcodeSteps.PROOF_VERIFIED:
                return (
                    <div className="flex flex-col items-center justify-center" style={{ width: size, height: size }}>
                        <div className="text-green-500 text-6xl mb-4">✓</div>
                        <div className="text-green-600 text-sm">Verification Successful</div>
                    </div>
                );

            default:
                return qrValue ? (
                    <QRCodeSVG
                        value={qrValue}
                        size={size}
                        bgColor={darkMode ? '#000000' : '#ffffff'}
                        fgColor={darkMode ? '#ffffff' : '#000000'}
                    />
                ) : (
                    <div className="flex items-center justify-center" style={{ width: size, height: size }}>
                        <div className="text-gray-500">Loading QR Code...</div>
                    </div>
                );
        }
    };

    return (
        <div className="angular-qr-wrapper flex flex-col items-center">
            {/* LED Status Indicator */}
            <div className="mb-2">
                <div
                    className="rounded-full transition-all duration-300"
                    style={{
                        width: '8px',
                        height: '8px',
                        backgroundColor: getLedColor(),
                        boxShadow: `0 0 12px ${getLedColor()}`,
                    }}
                />
            </div>

            {/* QR Code Container */}
            <div className="flex flex-col items-center justify-center">
                {renderContent()}
            </div>

            {/* Status Indicator */}
            <div className="mt-2 flex items-center space-x-2">
                <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
                <span className="text-xs text-gray-600">Angular SDK Active</span>
            </div>
        </div>
    );
};

export default AngularQRWrapper;
