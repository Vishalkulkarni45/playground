'use client';

import React, { useState, useEffect, useMemo } from 'react';
import dynamic from 'next/dynamic';
import { v4 as uuidv4 } from 'uuid';
import { countryCodes, SelfAppDisclosureConfig, type Country3LetterCode } from '@selfxyz/common';
import { countries, SelfAppBuilder } from '@selfxyz/qrcode';
import Image from 'next/image';
import type { SelfApp } from '@selfxyz/common';

// Import the QR code component with SSR disabled to prevent document references during server rendering
const SelfQRcodeWrapper = dynamic(
    () => import('@selfxyz/qrcode').then((mod) => mod.SelfQRcodeWrapper),
    { ssr: false }
);

function GoTestPlayground() {
    const [userId, setUserId] = useState<string | null>(null);
    const [savingOptions, setSavingOptions] = useState(false);
    const [selfApp, setSelfApp] = useState<SelfApp | null>(null);
    const [goServerStatus, setGoServerStatus] = useState<'checking' | 'connected' | 'error'>('checking');

    useEffect(() => {
        setUserId(uuidv4());
        checkGoServerStatus();
    }, []);

    const checkGoServerStatus = async () => {
        try {
            // Use OPTIONS request to go-saveOptions as a health check
            const response = await fetch('/api/go-saveOptions', {
                method: 'OPTIONS',
            });
            if (response.ok) {
                setGoServerStatus('connected');
            } else {
                setGoServerStatus('error');
            }
        } catch (error) {
            console.error('Go server connection error:', error);
            setGoServerStatus('error');
        }
    };

    const [disclosures, setDisclosures] = useState<SelfAppDisclosureConfig>({
        // DG1 disclosures
        issuing_state: false,
        name: false,
        nationality: true,
        date_of_birth: false,
        passport_number: false,
        gender: false,
        expiry_date: false,
        // Custom checks
        minimumAge: 18,
        excludedCountries: [
            countries.IRAN,
            countries.IRAQ,
            countries.NORTH_KOREA,
            countries.RUSSIA,
            countries.SYRIAN_ARAB_REPUBLIC,
            countries.VENEZUELA
        ] as Country3LetterCode[],
        ofac: true,
    });

    const [showCountryModal, setShowCountryModal] = useState(false);
    const [selectedCountries, setSelectedCountries] = useState<Country3LetterCode[]>([
        countries.IRAN,
        countries.IRAQ,
        countries.NORTH_KOREA,
        countries.RUSSIA,
        countries.SYRIAN_ARAB_REPUBLIC,
        countries.VENEZUELA
    ]);

    const [countrySelectionError, setCountrySelectionError] = useState<string | null>(null);
    const [searchQuery, setSearchQuery] = useState('');

    const handleAgeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const newAge = parseInt(e.target.value);
        setDisclosures(prev => ({ ...prev, minimumAge: newAge }));
    };

    const handleCountryToggle = (country: Country3LetterCode) => {
        setSelectedCountries(prev => {
            if (prev.includes(country)) {
                setCountrySelectionError(null);
                return prev.filter(c => c !== country);
            }

            if (prev.length >= 40) {
                setCountrySelectionError('Maximum 40 countries can be excluded');
                return prev;
            }

            return [...prev, country];
        });
    };

    const saveCountrySelection = () => {
        const codes = selectedCountries.map(countryName => {
            const entry = Object.entries(countryCodes).find(([_, name]) => name === countryName);
            return entry ? entry[0] : countryName.substring(0, 3).toUpperCase();
        }) as Country3LetterCode[];

        setDisclosures(prev => ({ ...prev, excludedCountries: codes }));
        setShowCountryModal(false);
    };

    const handleCheckboxChange = (field: string) => {
        setDisclosures(prev => ({
            ...prev,
            [field]: !prev[field as keyof typeof prev]
        }));
    };

    const saveOptionsToGoServer = useMemo(() => async () => {
        if (!userId || savingOptions || goServerStatus !== 'connected') return;

        setSavingOptions(true);
        try {
            const response = await fetch('/api/go-saveOptions', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    userId,
                    options: {
                        minimumAge: disclosures.minimumAge && disclosures.minimumAge > 0 ? disclosures.minimumAge : undefined,
                        excludedCountries: disclosures.excludedCountries,
                        ofac: disclosures.ofac,
                        issuing_state: disclosures.issuing_state,
                        name: disclosures.name,
                        nationality: disclosures.nationality,
                        date_of_birth: disclosures.date_of_birth,
                        passport_number: disclosures.passport_number,
                        gender: disclosures.gender,
                        expiry_date: disclosures.expiry_date
                    }
                }),
            });

            console.log("saved options to Go server");

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.message || 'Failed to save options to Go server');
            }
        } catch (error) {
            console.error('Error saving options to Go server:', error);
            if (error instanceof Error && error.message) {
                alert(error.message);
            } else {
                alert('Failed to save verification options to Go server. Please try again.');
            }
        } finally {
            setSavingOptions(false);
        }
    }, [userId, savingOptions, goServerStatus]);

    useEffect(() => {
        const timeoutId = setTimeout(() => {
            if (userId && goServerStatus === 'connected') {
                saveOptionsToGoServer();
            }
        }, 500);

        return () => clearTimeout(timeoutId);
    }, [userId, disclosures, goServerStatus]);

    useEffect(() => {
        if (userId && goServerStatus === 'connected') {
            const app = new SelfAppBuilder({
                appName: "Self Playground - Go Server Test",
                scope: "self-playground-go",
                endpoint: `${window.location.origin}/api/go-verify`,
                endpointType: "https", // HTTPS required for Self SDK
                logoBase64: "https://i.imgur.com/Rz8B3s7.png",
                userId,
                disclosures: {
                    ...disclosures,
                    minimumAge: disclosures.minimumAge && disclosures.minimumAge > 0 ? disclosures.minimumAge : undefined,
                },
                version: 2,
                userDefinedData: "hello from the Go server playground",
                devMode: true, // Enable dev mode for localhost
            } as Partial<SelfApp>).build();
            setSelfApp(app);
            console.log("selfApp built for Go server:", app);
        }
    }, [userId, disclosures, goServerStatus]);

    if (!userId) return null;

    const filteredCountries = Object.entries(countryCodes).filter(([_, country]) =>
        country.toLowerCase().includes(searchQuery.toLowerCase())
    );

    return (
        <div className="App flex flex-col min-h-screen bg-white text-black" suppressHydrationWarning>
            <nav className="w-full bg-white border-b border-gray-200 py-3 px-4 sm:px-6 flex items-center justify-between">
                <div className="flex items-center">
                    <div className="mr-4 sm:mr-8">
                        <Image
                            width={96}
                            height={36}
                            src="/self.svg"
                            alt="Self Logo"
                            className="h-9 w-24"
                        />
                    </div>
                    <div className="bg-blue-100 text-blue-800 px-3 py-1 rounded-md text-sm font-medium">
                        Go Server Test Mode
                    </div>
                </div>
                <div className="flex items-center space-x-2 sm:space-x-4">
                    <div className={`flex items-center space-x-2 px-3 py-1 rounded-md text-sm ${goServerStatus === 'connected' ? 'bg-green-100 text-green-800' :
                        goServerStatus === 'error' ? 'bg-red-100 text-red-800' :
                            'bg-yellow-100 text-yellow-800'
                        }`}>
                        <div className={`w-2 h-2 rounded-full ${goServerStatus === 'connected' ? 'bg-green-500' :
                            goServerStatus === 'error' ? 'bg-red-500' :
                                'bg-yellow-500'
                            }`}></div>
                        <span>
                            {goServerStatus === 'connected' ? 'Go Server Connected' :
                                goServerStatus === 'error' ? 'Go Server Error' :
                                    'Checking Go Server...'}
                        </span>
                    </div>
                    <button
                        onClick={checkGoServerStatus}
                        className="bg-blue-600 text-white px-3 py-1 rounded-md text-sm hover:bg-blue-700"
                    >
                        Refresh
                    </button>
                </div>
            </nav>

            {goServerStatus === 'error' && (
                <div className="bg-red-50 border-l-4 border-red-400 p-4 m-4">
                    <div className="flex">
                        <div className="ml-3">
                            <p className="text-sm text-red-700">
                                <strong>Go Server Connection Error:</strong> Vercel Go functions not responding
                            </p>
                            <p className="text-sm text-red-600 mt-1">
                                Check Vercel deployment status or try refreshing the page
                            </p>
                        </div>
                    </div>
                </div>
            )}

            <div className="flex-1 flex flex-col items-center justify-center px-4 py-8">
                <div className="w-full max-w-6xl flex flex-col md:flex-row gap-8">
                    <div className="w-full md:w-1/2 flex flex-col items-center justify-center">
                        {selfApp && goServerStatus === 'connected' ? (
                            <SelfQRcodeWrapper
                                selfApp={selfApp}
                                onSuccess={() => {
                                    console.log('Verification successful with Go server');
                                }}
                                darkMode={false}
                                onError={() => {
                                    console.error('Error generating QR code for Go server');
                                }}
                            />
                        ) : (
                            <p className="text-gray-500">
                                {goServerStatus === 'connected' ? 'Loading QR Code...' : 'Waiting for Go server connection...'}
                            </p>
                        )}
                        <p className="mt-4 text-sm text-gray-700">
                            User ID: {userId!.substring(0, 8)}...
                        </p>
                        <p className="text-xs text-blue-600 mt-1">
                            Backend: Go Server (Vercel)
                        </p>
                        <p className="text-xs text-gray-500 mt-1">
                            All APIs: Go Functions
                        </p>
                    </div>

                    <div className="w-full md:w-1/2 bg-white rounded-lg shadow-md p-6 border border-gray-300">
                        <h2 className="text-2xl font-semibold mb-4">Verification Options</h2>

                        <div className="space-y-6">
                            <div className="border border-gray-300 rounded-md p-4">
                                <h3 className="text-lg font-medium mb-3">Personal Information</h3>
                                <div className="grid grid-cols-2 gap-4">
                                    <div className="space-y-2">
                                        <label className="flex items-center space-x-2">
                                            <input
                                                type="checkbox"
                                                checked={disclosures.issuing_state}
                                                onChange={() => handleCheckboxChange('issuing_state')}
                                                className="h-4 w-4"
                                            />
                                            <span>Disclose Issuing State</span>
                                        </label>
                                        <label className="flex items-center space-x-2">
                                            <input
                                                type="checkbox"
                                                checked={disclosures.name}
                                                onChange={() => handleCheckboxChange('name')}
                                                className="h-4 w-4"
                                            />
                                            <span>Disclose Name</span>
                                        </label>
                                        <label className="flex items-center space-x-2">
                                            <input
                                                type="checkbox"
                                                checked={disclosures.nationality}
                                                onChange={() => handleCheckboxChange('nationality')}
                                                className="h-4 w-4"
                                            />
                                            <span>Disclose Nationality</span>
                                        </label>
                                        <label className="flex items-center space-x-2">
                                            <input
                                                type="checkbox"
                                                checked={disclosures.date_of_birth}
                                                onChange={() => handleCheckboxChange('date_of_birth')}
                                                className="h-4 w-4"
                                            />
                                            <span>Disclose Date of Birth</span>
                                        </label>
                                    </div>
                                    <div className="space-y-2">
                                        <label className="flex items-center space-x-2">
                                            <input
                                                type="checkbox"
                                                checked={disclosures.passport_number}
                                                onChange={() => handleCheckboxChange('passport_number')}
                                                className="h-4 w-4"
                                            />
                                            <span>Disclose Passport Number</span>
                                        </label>
                                        <label className="flex items-center space-x-2">
                                            <input
                                                type="checkbox"
                                                checked={disclosures.gender}
                                                onChange={() => handleCheckboxChange('gender')}
                                                className="h-4 w-4"
                                            />
                                            <span>Disclose Gender</span>
                                        </label>
                                        <label className="flex items-center space-x-2">
                                            <input
                                                type="checkbox"
                                                checked={disclosures.expiry_date}
                                                onChange={() => handleCheckboxChange('expiry_date')}
                                                className="h-4 w-4"
                                            />
                                            <span>Disclose Expiry Date</span>
                                        </label>
                                    </div>
                                </div>
                            </div>
                            <div className="border border-gray-300 rounded-md p-4">
                                <h3 className="text-lg font-medium mb-3">Verification Rules</h3>
                                <div className="space-y-4">
                                    <div>
                                        <label className="block mb-1">Minimum Age: {disclosures.minimumAge || 'None'}</label>
                                        <input
                                            type="range"
                                            min="0"
                                            max="99"
                                            value={disclosures.minimumAge}
                                            onChange={handleAgeChange}
                                            className="w-full"
                                        />
                                        <div className="text-sm text-gray-500 mt-1">
                                            Set to 0 to disable age requirement
                                        </div>
                                    </div>
                                    <div>
                                        <label className="flex items-center space-x-2">
                                            <input
                                                type="checkbox"
                                                checked={disclosures.ofac}
                                                onChange={() => handleCheckboxChange('ofac')}
                                                className="h-4 w-4"
                                            />
                                            <span>Enable OFAC Check</span>
                                        </label>
                                    </div>

                                    <div>
                                        <button
                                            onClick={() => setShowCountryModal(true)}
                                            className="px-3 py-1 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
                                        >
                                            Configure Excluded Countries
                                        </button>
                                        <div className="mt-2 text-sm text-gray-700">
                                            {disclosures.excludedCountries && disclosures.excludedCountries.length > 0
                                                ? `${disclosures.excludedCountries.length} countries excluded`
                                                : "No countries excluded"}
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            {/* Country Selection Modal */}
            {showCountryModal && (
                <div className="fixed inset-0 bg-white bg-opacity-90 flex items-center justify-center z-50">
                    <div className="bg-white rounded-lg shadow-xl p-6 max-w-2xl w-full max-h-[80vh] overflow-y-auto border border-gray-300">
                        <h3 className="text-xl font-semibold mb-4">Select Countries to Exclude</h3>

                        {countrySelectionError && (
                            <div className="mb-4 p-2 bg-red-100 text-red-700 rounded">
                                {countrySelectionError}
                            </div>
                        )}

                        <div className="mb-4">
                            <input
                                type="text"
                                placeholder="Search countries..."
                                value={searchQuery}
                                onChange={(e) => setSearchQuery(e.target.value)}
                                className="w-full p-2 border border-gray-300 rounded bg-white text-black"
                            />
                        </div>

                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-2 mb-6 max-h-80 overflow-y-auto">
                            {filteredCountries.map(([code, country]) => (
                                <label key={code} className="flex items-center space-x-2 p-1 hover:bg-gray-100 rounded">
                                    <input
                                        type="checkbox"
                                        checked={selectedCountries.includes(code as Country3LetterCode)}
                                        onChange={() => handleCountryToggle(code as Country3LetterCode)}
                                        className="h-4 w-4"
                                    />
                                    <span className="text-sm">{country}</span>
                                </label>
                            ))}
                        </div>

                        <div className="flex justify-end space-x-3">
                            <button
                                onClick={() => setShowCountryModal(false)}
                                className="px-4 py-2 border border-gray-300 rounded hover:bg-gray-100"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={saveCountrySelection}
                                className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
                            >
                                Apply
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}

export default GoTestPlayground;
